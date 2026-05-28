# Architecture

UniRTM's engine is written purely in Go, utilizing goroutines for extreme parallel downloading. It does not spawn subshells for simple operations, retaining memory efficiency.