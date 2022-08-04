TODO
- Make a real readme :)
- Use a real license
- Temp caching as implemented sucks, it remembers the last CPU temp, however a 1C variance will still cause the fans to require updating. Possibly need to cache values at last update so we can allow for minor temperature fluxtuations. Note: caching is at top of mind here as I suspect that I'm crashing the BMC by updating the fan speed too often. It may be worthwhile being a little more aggressive on speeds so we can be more passive on the frequency of speed updates.
- On the above note, maybe an interim mitigation of ignoring temp reductions?
- Re-org code so this can be a real go package
- Implement real daemon()izing code
- Change up logging to a proper logging framework so stdout isn't where all logs go
- Probably should capture the output of ipmitool (if any) for logging if it has an error
- ipmitool.sh can probabl go now, was a useful debugging tool that is no longer needed

Current console output
Found 2 CPU packages
Current hottest CPU at 52C
Setting fan speed to 44%
Found 2 CPU packages
Current hottest CPU at 52C
Found 2 CPU packages
Current hottest CPU at 50C
Setting fan speed to 40%
Found 2 CPU packages
Current hottest CPU at 50C
Found 2 CPU packages
