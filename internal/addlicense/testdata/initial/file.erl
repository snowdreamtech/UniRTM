% Copyright (c) 2026 SnowdreamTech. All rights reserved.
% Licensed under the MIT License. See LICENSE file in the project root for full license information.

-module(hello).
-export([hello_world/0]).

hello_world() -> io:fwrite("hello, world\n").
