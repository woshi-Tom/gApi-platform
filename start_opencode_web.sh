#!/bin/bash
pkill -9 opencode

/root/.opencode/bin/opencode web --hostname 0.0.0.0 --port 4096

