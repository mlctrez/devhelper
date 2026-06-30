#!/bin/bash
curl -s http://localhost:9990/quit 2>/dev/null
sleep 1
go run .
