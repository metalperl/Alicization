# Cron KPIs Input Plugin

This plugin gathers the KPI data from Cron jobs

* Unique User Daily Count

### Configuration

```toml
# Create KPI's from Cron jobs
[[inputs.cron_kpis]]
  ## Location of the cron file log
  ##   For example:
  ##    "/home/naveen/go/src/filehandling/test.txt"
  location = "/var/log/syslog"

  ## The name of the cron job to target
  ##   For example:
  ##     "WatchDogTimer.check"
  cronjob  = "WatchDogTimer.check"

  ## Host where the Cron jobs are running
  host     = "droozy-den-1p"

  ## Cron Count, the count of unique jobs that run for each day
  ##    jobs that run multiple times in a day are still only counted once
  ##    counts are listed starting with Sunday and need to be input for each day
  cron_count = ["3","4","5","6","7","0","0"]
```

### Metrics:
* Cron Job successfull run - numeric of cron jobs for that day (value for day or 0, 0 = cron not running)


## Tags
* All measurements has following tags
    * host (the host name from which the metrics are gathered)

## Fields
    * cron_count: Unique sount of crons for that day


### Example Output:

This section shows example output in Line Protocol format.  You can often use
`telegraf --input-filter cron_kpis --test` or use the `file` output to get
this information.

```toml
cron_kpis map[cron_count:44] map[host:HostName]
```