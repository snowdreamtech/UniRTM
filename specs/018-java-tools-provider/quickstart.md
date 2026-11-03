# Quickstart Validation Guide

## Validating Maven Provider Download and Shimming

1. **Prerequisites**: Ensure `java` is installed and available in `PATH`.
2. **Execute Installation**:

   ```bash
   unirtm use -g maven:com.puppycrawl.tools/checkstyle@10.15.0
   ```

3. **Verify Execution**:

   ```bash
   checkstyle --version
   ```

   *Expected Outcome:* The checkstyle version is printed to stdout.

4. **Verify Shim Context**:

   ```bash
   cat ~/.unirtm/installs/maven/com.puppycrawl.tools/checkstyle/10.15.0/bin/checkstyle
   ```

   *Expected Outcome:* The shim contains `exec java -jar .../checkstyle-10.15.0.jar "$@"` (on Unix).
