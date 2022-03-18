# TunnelEffect long term testing utility

Telttest checks TunnelEffect operations when it is configured to copy
files from one folder to another.
This utility continuously generats malicious and legitimate files
and checks that each one of the files is either copied to 
target folder or to Quarantine.

## Building

Use ```go build``` command from project root directory to build telttest utility.

## Running

Create telttest.yaml file with following content:
```
SourceDir: <TunnelEffect source folder - according to tunnelfeect.yaml configuration>
TargetDir: <TunnelEffet target folder - according to tunnelfeect.yaml configuration>
QuarantineDir: <TunnelEffect quarantine folder - according to tunnelfeect.yaml configuration>
Log: telttest.log (log file name and path)
Timeout: 20m (time to wait for copying to target or moving quarantine)
Pause: 5m (time to wait between sample generation)
```

Run ```./telttest``` and monitor telttest.log file for errors.

