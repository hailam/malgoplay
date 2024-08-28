# MalgoPlay

## Table of Contents

1. [Introduction](#introduction)
2. [Command-line Arguments](#command-line-arguments)

## Introduction

The malgoplay is a Go program that generates and analyzes sine waves. It can produce a fixed frequency or sweep through a range of frequencies. The program also detects the frequency of the generated sound and provides real-time feedback on the match between the generated and detected frequencies.

## Command-line Arguments

- `-f`, `--frequency`: Maximum frequency of the sine wave in Hz (default: 1000)
- `-m`, `--min-frequency`: Minimum frequency to start sweeping from in Hz (default: 0, which means no sweeping)
- `-a`, `--amplitude`: Amplitude of the sine wave (default: 0.5)
- `-r`, `--sample-rate`: Sample rate in Hz (default: 48000)
- `-s`, `--sweep-rate`: Frequency change rate in Hz per second when sweeping (default: 1.0)

## Examples

1. Generate a fixed 440 Hz tone:

   ```
   {{binary / go run .}} -f 440
   ```

2. Sweep from 100 Hz to 1000 Hz, changing by 10 Hz per second:

   ```
   {{binary / go run .}} -f 1000 -m 100 -s 10
   ```
