// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

group "default" {
  targets = ["build"]
}

target "build" {
  dockerfile = "./Dockerfile"
  output = ["type=docker"]
}
