// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RegisterPortsRoutes registers port forwarding and detection routes.
func RegisterPortsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/ports/list", handlePortsList)
	mux.HandleFunc("/api/ports/forward", handlePortForward)
	mux.HandleFunc("/api/ports/update", handlePortUpdate)
	mux.HandleFunc("/api/ports/stop", handlePortStop)
	mux.HandleFunc("/api/ports/unforward", handlePortUnforward)
	mux.HandleFunc("/api/ports/preview/", handlePortPreview)
	mux.HandleFunc("/api/ports/probe", handlePortProbe)
	mux.HandleFunc("/api/ports/workspace-hints", handlePortsWorkspaceHints)
}

type PortProbeInfo struct {
	Port            int    `json:"port"`
	IsHTTPS         bool   `json:"isHttps"`
	HTTPStatus      int    `json:"httpStatus,omitempty"`
	ServerHeader    string `json:"serverHeader,omitempty"`
	XPoweredBy      string `json:"xPoweredBy,omitempty"`
	Location        string `json:"location,omitempty"`
	DetectedService string `json:"detectedService,omitempty"`
	ProbeDurationMs int64  `json:"probeDurationMs,omitempty"`
}

type WorkspaceServiceHint struct {
	Port        int    `json:"port"`
	Name        string `json:"name"`
	SourceFile  string `json:"sourceFile"`
	ServiceType string `json:"serviceType"`
	Framework   string `json:"framework,omitempty"`
}

type SystemListeningPort struct {
	Port         int            `json:"port"`
	Protocol     string         `json:"protocol"`
	LocalAddress string         `json:"localAddress"`
	Pid          int            `json:"pid,omitempty"`
	Process      string         `json:"process,omitempty"`
	CommandLine  string         `json:"commandLine,omitempty"`
	Probe        *PortProbeInfo `json:"probe,omitempty"`
}

type PortForwardRequest struct {
	Port       int    `json:"port"`
	LocalPort  int    `json:"localPort,omitempty"`
	Label      string `json:"label,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	Visibility string `json:"visibility,omitempty"`
}

var (
	portsCacheMu   sync.RWMutex
	portsCacheData []SystemListeningPort
	portsCacheTime time.Time
	portsCacheTTL  = 1500 * time.Millisecond
)

// InvalidatePortsCache clears the listening ports cache.
func InvalidatePortsCache() {
	portsCacheMu.Lock()
	portsCacheTime = time.Time{}
	portsCacheMu.Unlock()
}

func scanListeningPorts() ([]SystemListeningPort, error) {
	portsCacheMu.RLock()
	if time.Since(portsCacheTime) < portsCacheTTL && portsCacheData != nil {
		cached := make([]SystemListeningPort, len(portsCacheData))
		copy(cached, portsCacheData)
		portsCacheMu.RUnlock()
		return cached, nil
	}
	portsCacheMu.RUnlock()

	portsMap := make(map[int]SystemListeningPort)

	// 1. Try ss (fastest, standard on modern Linux)
	if ssPath, err := exec.LookPath("ss"); err == nil {
		out, err := exec.Command(ssPath, "-tlpn").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if !strings.HasPrefix(line, "LISTEN") {
					continue
				}
				fields := strings.Fields(line)
				if len(fields) < 4 {
					continue
				}
				localAddr := fields[3]
				port := extractPortFromAddr(localAddr)
				if port <= 0 || port > 65535 {
					continue
				}

				item := SystemListeningPort{
					Port:         port,
					Protocol:     "tcp",
					LocalAddress: localAddr,
				}

				if len(fields) >= 5 {
					procField := strings.Join(fields[4:], " ")
					pid := extractPidFromUsers(procField)
					if pid > 0 {
						item.Pid = pid
						item.CommandLine = readProcessCmdline(pid)
						item.Process = extractProcessNameFromUsers(procField)
					}
				}
				portsMap[port] = item
			}
		}
	}

	// 2. Try lsof as supplementary / fallback (useful for macOS or when ss couldn't resolve ports)
	if len(portsMap) == 0 || runtime.GOOS != "linux" {
		if lsofPath, err := exec.LookPath("lsof"); err == nil {
			out, err := exec.Command(lsofPath, "-iTCP", "-sTCP:LISTEN", "-P", "-n").Output()
			if err == nil {
				lines := strings.Split(string(out), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "COMMAND") || line == "" {
						continue
					}
					fields := strings.Fields(line)
					if len(fields) < 9 {
						continue
					}
					procName := fields[0]
					pid, _ := strconv.Atoi(fields[1])
					nameField := fields[8]
					port := extractPortFromAddr(nameField)
					if port <= 0 || port > 65535 {
						continue
					}

					existing, ok := portsMap[port]
					if !ok {
						existing = SystemListeningPort{
							Port:         port,
							Protocol:     "tcp",
							LocalAddress: nameField,
						}
					}
					if existing.Pid == 0 && pid > 0 {
						existing.Pid = pid
						existing.Process = procName
						existing.CommandLine = readProcessCmdline(pid)
					}
					portsMap[port] = existing
				}
			}
		}
	}

	// 3. Try netstat (standard fallback for Windows: netstat -ano -p tcp)
	if netstatPath, err := exec.LookPath("netstat"); err == nil {
		netstatArgs := []string{"-ano", "-p", "tcp"}
		if runtime.GOOS != "windows" {
			netstatArgs = []string{"-tuln"}
		}
		out, err := exec.Command(netstatPath, netstatArgs...).Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if !strings.Contains(line, "LISTEN") {
					continue
				}
				fields := strings.Fields(line)
				if len(fields) < 4 {
					continue
				}
				localAddr := fields[1]
				if runtime.GOOS != "windows" && len(fields) >= 4 {
					localAddr = fields[3]
				}
				port := extractPortFromAddr(localAddr)
				if port <= 0 || port > 65535 {
					continue
				}

				if _, exists := portsMap[port]; !exists {
					item := SystemListeningPort{
						Port:         port,
						Protocol:     "tcp",
						LocalAddress: localAddr,
					}
					if len(fields) >= 5 {
						lastField := fields[len(fields)-1]
						if pid, err := strconv.Atoi(lastField); err == nil && pid > 0 {
							item.Pid = pid
							item.CommandLine = readProcessCmdline(pid)
						}
					}
					portsMap[port] = item
				}
			}
		}
	}

	result := make([]SystemListeningPort, 0, len(portsMap))
	for _, item := range portsMap {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Port < result[j].Port
	})

	portsCacheMu.Lock()
	portsCacheData = result
	portsCacheTime = time.Now()
	portsCacheMu.Unlock()

	return result, nil
}

func extractPortFromAddr(addr string) int {
	addr = strings.TrimSpace(addr)
	if idx := strings.LastIndex(addr, ":"); idx != -1 && idx < len(addr)-1 {
		portStr := addr[idx+1:]
		if p, err := strconv.Atoi(portStr); err == nil {
			return p
		}
	}
	return 0
}

func extractPidFromUsers(s string) int {
	idx := strings.Index(s, "pid=")
	if idx == -1 {
		return 0
	}
	sub := s[idx+4:]
	end := strings.IndexAny(sub, ",)")
	if end != -1 {
		sub = sub[:end]
	}
	p, _ := strconv.Atoi(sub)
	return p
}

func extractProcessNameFromUsers(s string) string {
	idx := strings.Index(s, "((\"")
	if idx != -1 {
		sub := s[idx+3:]
		end := strings.Index(sub, "\"")
		if end != -1 {
			return sub[:end]
		}
	}
	return ""
}

func readProcessCmdline(pid int) string {
	// 1. Linux: /proc/<pid>/cmdline
	cmdPath := fmt.Sprintf("/proc/%d/cmdline", pid)
	data, err := os.ReadFile(cmdPath)
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(strings.ReplaceAll(string(data), "\x00", " "))
	}

	// 2. macOS / BSD: ps -p <pid> -o command=
	if runtime.GOOS == "darwin" {
		out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=").Output()
		if err == nil && len(out) > 0 {
			return strings.TrimSpace(string(out))
		}
	}

	// 3. Windows: tasklist /FI "PID eq <pid>" /FO CSV /NH
	if runtime.GOOS == "windows" {
		out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH").Output()
		if err == nil && len(out) > 0 {
			parts := strings.Split(string(out), ",")
			if len(parts) > 0 {
				return strings.Trim(parts[0], "\" \r\n")
			}
		}
	}

	return ""
}

func probePort(port int) PortProbeInfo {
	start := time.Now()
	info := PortProbeInfo{
		Port:    port,
		IsHTTPS: false,
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	// 1. Try TLS probe first (<= 35ms timeout)
	tlsDialer := &net.Dialer{Timeout: 35 * time.Millisecond}
	tlsConn, err := tls.DialWithDialer(tlsDialer, "tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err == nil {
		info.IsHTTPS = true
		_ = tlsConn.SetDeadline(time.Now().Add(25 * time.Millisecond))
		reqStr := fmt.Sprintf("HEAD / HTTP/1.1\r\nHost: 127.0.0.1:%d\r\nUser-Agent: UniRTM-Probe\r\nConnection: close\r\n\r\n", port)
		_, _ = tlsConn.Write([]byte(reqStr))
		buf := make([]byte, 1024)
		n, _ := tlsConn.Read(buf)
		_ = tlsConn.Close()
		if n > 0 {
			parseHTTPProbeResponse(string(buf[:n]), &info)
		}
		info.ProbeDurationMs = time.Since(start).Milliseconds()
		return info
	}

	// 2. Plain TCP probe (<= 35ms timeout)
	conn, err := net.DialTimeout("tcp", addr, 35*time.Millisecond)
	if err != nil {
		info.ProbeDurationMs = time.Since(start).Milliseconds()
		return info
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(25 * time.Millisecond))
	reqStr := fmt.Sprintf("HEAD / HTTP/1.1\r\nHost: 127.0.0.1:%d\r\nUser-Agent: UniRTM-Probe\r\nConnection: close\r\n\r\n", port)
	_, _ = conn.Write([]byte(reqStr))

	buf := make([]byte, 1024)
	n, readErr := conn.Read(buf)
	if readErr == nil && n > 0 {
		respStr := string(buf[:n])
		if strings.HasPrefix(respStr, "HTTP/") {
			parseHTTPProbeResponse(respStr, &info)
			info.ProbeDurationMs = time.Since(start).Milliseconds()
			return info
		}
		detectNonHTTPProtocols(respStr, &info)
		if info.DetectedService != "" {
			info.ProbeDurationMs = time.Since(start).Milliseconds()
			return info
		}
	}

	// 3. If no HTTP response or port is 6379, try Redis PING
	if port == 6379 || info.DetectedService == "" {
		if rConn, err := net.DialTimeout("tcp", addr, 25*time.Millisecond); err == nil {
			_ = rConn.SetDeadline(time.Now().Add(20 * time.Millisecond))
			_, _ = rConn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
			rBuf := make([]byte, 256)
			rn, rErr := rConn.Read(rBuf)
			_ = rConn.Close()
			if rErr == nil && rn > 0 {
				rStr := string(rBuf[:rn])
				if strings.HasPrefix(rStr, "+PONG") || strings.Contains(rStr, "NOAUTH") || strings.Contains(rStr, "ERR") {
					info.DetectedService = "Redis"
				}
			}
		}
	}

	info.ProbeDurationMs = time.Since(start).Milliseconds()
	return info
}

func parseHTTPProbeResponse(respStr string, info *PortProbeInfo) {
	lines := strings.Split(respStr, "\r\n")
	if len(lines) == 0 {
		return
	}
	statusLine := lines[0]
	statusParts := strings.Fields(statusLine)
	if len(statusParts) >= 2 {
		if code, err := strconv.Atoi(statusParts[1]); err == nil {
			info.HTTPStatus = code
		}
	}

	for _, line := range lines[1:] {
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			k := strings.ToLower(strings.TrimSpace(parts[0]))
			v := strings.TrimSpace(parts[1])
			switch k {
			case "server":
				info.ServerHeader = v
			case "x-powered-by":
				info.XPoweredBy = v
			case "location":
				info.Location = v
			}
		}
	}

	serverLower := strings.ToLower(info.ServerHeader)
	xPoweredLower := strings.ToLower(info.XPoweredBy)

	if strings.Contains(serverLower, "uvicorn") {
		info.DetectedService = "FastAPI / Starlette (Uvicorn)"
	} else if strings.Contains(serverLower, "werkzeug") {
		info.DetectedService = "Flask (Werkzeug)"
	} else if strings.Contains(serverLower, "gunicorn") {
		info.DetectedService = "Gunicorn WSGI"
	} else if strings.Contains(xPoweredLower, "express") {
		info.DetectedService = "Express (Node.js)"
	} else if strings.Contains(xPoweredLower, "next.js") {
		info.DetectedService = "Next.js"
	} else if strings.Contains(serverLower, "nginx") {
		info.DetectedService = "Nginx Proxy"
	} else if strings.Contains(serverLower, "caddy") {
		info.DetectedService = "Caddy Server"
	} else if strings.Contains(serverLower, "apache") {
		info.DetectedService = "Apache HTTP Server"
	} else if strings.Contains(serverLower, "kestrel") {
		info.DetectedService = "ASP.NET Core (Kestrel)"
	} else if info.ServerHeader != "" {
		info.DetectedService = info.ServerHeader
	} else if info.XPoweredBy != "" {
		info.DetectedService = info.XPoweredBy
	}
}

func detectNonHTTPProtocols(data string, info *PortProbeInfo) {
	if strings.Contains(data, "mysql_native_password") || strings.Contains(data, "caching_sha2_password") || strings.Contains(data, "MariaDB") {
		info.DetectedService = "MySQL / MariaDB"
	} else if strings.Contains(data, "SSH-") {
		info.DetectedService = "SSH Server"
	}
}

func handlePortsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ports, err := scanListeningPorts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	if r.URL.Query().Get("probe") == "true" {
		var wg sync.WaitGroup
		for i := range ports {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				probeRes := probePort(ports[idx].Port)
				ports[idx].Probe = &probeRes
			}(i)
		}
		wg.Wait()
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"listening": ports,
	})
}

func handlePortProbe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Port int `json:"port"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Port <= 0 || req.Port > 65535 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid port"})
		return
	}
	res := probePort(req.Port)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"probe":   res,
	})
}

func handlePortForward(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req PortForwardRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Port <= 0 || req.Port > 65535 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid port"})
		return
	}
	InvalidatePortsCache()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"port":    req.Port,
	})
}

func handlePortUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req PortForwardRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Port <= 0 || req.Port > 65535 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid port"})
		return
	}
	InvalidatePortsCache()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"port":    req.Port,
	})
}

func handlePortStop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req PortForwardRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	InvalidatePortsCache()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"port":    req.Port,
	})
}

func handlePortUnforward(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req PortForwardRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	InvalidatePortsCache()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"port":    req.Port,
	})
}

func handlePortPreview(w http.ResponseWriter, r *http.Request) {
	subPath := strings.TrimPrefix(r.URL.Path, "/api/ports/preview/")
	parts := strings.SplitN(subPath, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Port required in path", http.StatusBadRequest)
		return
	}
	port, err := strconv.Atoi(parts[0])
	if err != nil || port <= 0 || port > 65535 {
		http.Error(w, "Invalid port number", http.StatusBadRequest)
		return
	}

	targetURL, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	remainder := "/"
	if len(parts) > 1 {
		remainder = "/" + parts[1]
	}
	r.URL.Path = remainder

	proxy.ServeHTTP(w, r)
}

func scanWorkspaceServiceHints(rootDir string) []WorkspaceServiceHint {
	hints := make([]WorkspaceServiceHint, 0)
	portMap := make(map[int]WorkspaceServiceHint)

	if rootDir == "" {
		rootDir, _ = os.Getwd()
	}

	// 1. Scan package.json in root and frontend subfolder
	pkgFiles := []string{
		filepath.Join(rootDir, "package.json"),
		filepath.Join(rootDir, "frontend", "package.json"),
	}
	portRegex := regexp.MustCompile(`--port\s+(\d+)|-p\s+(\d+)|PORT=(\d+)`)
	for _, pkgPath := range pkgFiles {
		data, err := os.ReadFile(pkgPath)
		if err == nil {
			var pkg struct {
				Name         string            `json:"name"`
				Scripts      map[string]string `json:"scripts"`
				Dependencies map[string]string `json:"dependencies"`
			}
			if json.Unmarshal(data, &pkg) == nil {
				fw := "Node.js"
				if _, ok := pkg.Dependencies["next"]; ok {
					fw = "Next.js"
				} else if _, ok := pkg.Dependencies["vite"]; ok {
					fw = "Vite"
				} else if _, ok := pkg.Dependencies["vue"]; ok {
					fw = "Vue"
				} else if _, ok := pkg.Dependencies["react"]; ok {
					fw = "React"
				}

				for sName, sCmd := range pkg.Scripts {
					matches := portRegex.FindStringSubmatch(sCmd)
					if len(matches) > 0 {
						for _, m := range matches[1:] {
							if m != "" {
								if p, err := strconv.Atoi(m); err == nil && p > 0 && p <= 65535 {
									name := fmt.Sprintf("%s (%s:%s)", pkg.Name, filepath.Base(pkgPath), sName)
									if pkg.Name == "" {
										name = fmt.Sprintf("Web App (%s:%s)", filepath.Base(pkgPath), sName)
									}
									portMap[p] = WorkspaceServiceHint{
										Port:        p,
										Name:        name,
										SourceFile:  filepath.Base(pkgPath),
										ServiceType: "npm-script",
										Framework:   fw,
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// 2. Scan compose files
	composeFiles := []string{
		filepath.Join(rootDir, "docker-compose.yml"),
		filepath.Join(rootDir, "docker-compose.yaml"),
		filepath.Join(rootDir, "compose.yml"),
		filepath.Join(rootDir, "compose.yaml"),
	}
	for _, cPath := range composeFiles {
		data, err := os.ReadFile(cPath)
		if err == nil {
			lines := strings.Split(string(data), "\n")
			currentService := ""
			inPorts := false
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
					continue
				}
				if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") && strings.HasSuffix(trimmed, ":") {
					currentService = strings.TrimSuffix(trimmed, ":")
					inPorts = false
					continue
				}
				if trimmed == "ports:" {
					inPorts = true
					continue
				}
				if inPorts && strings.HasPrefix(trimmed, "-") {
					val := strings.Trim(strings.TrimPrefix(trimmed, "-"), " \"'")
					parts := strings.Split(val, ":")
					var hostPortStr string
					if len(parts) == 2 {
						hostPortStr = parts[0]
					} else if len(parts) == 3 {
						hostPortStr = parts[1]
					}
					if hostPortStr != "" {
						if p, err := strconv.Atoi(hostPortStr); err == nil && p > 0 && p <= 65535 {
							fw := "Docker Service"
							srvLower := strings.ToLower(currentService)
							if strings.Contains(srvLower, "postgres") || strings.Contains(srvLower, "db") {
								fw = "PostgreSQL"
							} else if strings.Contains(srvLower, "redis") {
								fw = "Redis"
							} else if strings.Contains(srvLower, "mysql") {
								fw = "MySQL"
							} else if strings.Contains(srvLower, "mongo") {
								fw = "MongoDB"
							}
							portMap[p] = WorkspaceServiceHint{
								Port:        p,
								Name:        fmt.Sprintf("%s (compose: %s)", currentService, currentService),
								SourceFile:  filepath.Base(cPath),
								ServiceType: "compose",
								Framework:   fw,
							}
						}
					}
				} else if inPorts && !strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "\t\t") {
					inPorts = false
				}
			}
		}
	}

	// 3. Scan .env files
	envFiles := []string{
		filepath.Join(rootDir, ".env"),
		filepath.Join(rootDir, ".env.local"),
		filepath.Join(rootDir, ".env.development"),
	}
	for _, envPath := range envFiles {
		data, err := os.ReadFile(envPath)
		if err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				k := strings.ToUpper(strings.TrimSpace(parts[0]))
				v := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
				if strings.Contains(k, "PORT") {
					if p, err := strconv.Atoi(v); err == nil && p > 0 && p <= 65535 {
						if _, exists := portMap[p]; !exists {
							portMap[p] = WorkspaceServiceHint{
								Port:        p,
								Name:        fmt.Sprintf("Env Service (%s)", k),
								SourceFile:  filepath.Base(envPath),
								ServiceType: "env-var",
								Framework:   "Custom",
							}
						}
					}
				}
			}
		}
	}

	for _, item := range portMap {
		hints = append(hints, item)
	}
	sort.Slice(hints, func(i, j int) bool {
		return hints[i].Port < hints[j].Port
	})

	return hints
}

func handlePortsWorkspaceHints(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pwd, _ := os.Getwd()
	hints := scanWorkspaceServiceHints(pwd)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"hints":   hints,
	})
}
