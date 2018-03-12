#!/bin/bash

ps x | grep zipper | awk '{print $1}' | xargs kill >null 2>&1
rm -f null