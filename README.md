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

## Environment


v, ok := os.LookupEnv(p("SOURCE_DIR"))
	if ok {
		c.SourceDir = v
	}
	v, ok = os.LookupEnv(p("TARGET_DIR"))
	if ok {
		c.TargetDir = v
	}
	v, ok = os.LookupEnv(p("QUARANTINE_DIR"))
	if ok {
		c.QuarantineDir = v
	}
	v, ok = os.LookupEnv(p("LOG"))
	if ok {
		c.Log = v
	}
	v, ok = os.LookupEnv(p("TIMEOUT"))
	if ok {
		var err error
		c.Timeout, err = time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("%s: %w", p("TIMEOUT"), err)
		}
	}
	v, ok = os.LookupEnv(p("PAUSE"))
	if ok {
		var err error
		c.Pause, err = time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("%s: %w", p("PAUSE"), err)
		}
	}