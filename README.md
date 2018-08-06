# Jira Input Plugin

This plugin gathers the jira's from JIRA server

* Weekly Jira statuses
* Bi-Weekly Jira statuses
* Monthly Jira statuses
* Quarterly Jira statuses
* Yearly Jira statuses


### Configuration

```toml
# Read metrics from one or many jira servers
[[inputs.jira]]
  ## This plugin will query supplied project in jira
  ## Jira server to connect to
  ##  [protocol://[(hostname)]]
  ##  e.g.
  ##    https://jira.com
  ##
  ## if no servers are specified, local machine will be queried
  ##
  server = [http://127.0.0.1:8080"]

  ## JIRA Project
  project = ""

  ## HTTP Basic Authentication username and password.
  username = ""
  password = ""

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "json"

  ## gather metrics from issues within the last week for the jira server provided above
  gather_weekly    = false
  #
  ## gather metrics from issues within the last 2 weeks for the jira server provided above
  gather_biweekly  = true
  #
  ## gather metrics from issues within the month for the jira server provided above
  gather_monthly   = true
  #
  ## gather metrics from issues opened within the quarter for the jira server provided above
  gather_quarterly = false
  #
  ## gather metrics from issues opened within the year for the jira server provided above
  gather_yearly    = false
```

### Metrics:

* Weekly statuses - all opened (open typed) and closed (closed typed) values in jira project for `WEEKLY STATUSES`
* BiWeekly statuses - all opened (open typed) and closed (closed typed) values in jira project for `BIWEEKLY STATUSES`
* Monthly statuses - all opened (open typed) and closed (closed typed) values in jira project for `MONTHLY STATUSES`
* Quarterly statuses - all opened (open typed) and closed (closed typed) values in jira project for `QUARTERLY STATUSES`
* Yearly statuses - all opened (open typed) and closed (closed typed) values in jira project for `YEARLY STATUSES`


## Tags

* All measurements have following tags
    * project:      Project
    * epoch:        epoch

## Fields
    * opened_jiras: opnd
    * closed_jiras: clsd


### Sample Queries:

This section should contain some useful InfluxDB queries that can be used to
get started with the plugin or to generate dashboards.  For each query listed,
describe at a high level what data is returned.

Get the max, mean, and min for the measurement in the last hour:
```toml
SELECT max(field1), mean(field1), min(field1) FROM measurement1 WHERE tag1=bar AND time > now() - 1h GROUP BY tag
```


### Example Output:

This section shows example output in Line Protocol format.  You can often use
`telegraf --input-filter jira_kpis --test` or use the `file` output to get
this information.

```toml
measurement1,tag1=foo,tag2=bar field1=1i,field2=2.1 1453831884664956455
measurement2,tag1=foo,tag2=bar,tag3=baz field3=1i 1453831884664956455
```