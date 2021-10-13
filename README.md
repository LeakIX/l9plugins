# l9 suite stock plugins

[![GitHub Release](https://img.shields.io/github/v/release/LeakIX/l9plugins)](https://github.com/LeakIX/l9plugins/releases)
[![Follow on Twitter](https://img.shields.io/twitter/follow/leak_ix.svg?logo=twitter)](https://twitter.com/leak_ix)

This repository contains LeakIX maintained plugins implementing the [l9format golang plugin interface](https://github.com/LeakIX/l9format/blob/master/l9plugin.go).
They are currently used by [l9explore](https://github.com/LeakIX/l9explore) but could be implemented by Go security tool.

## Current plugins

|Plugin|Protocols|Stage|Description|Author|
|------|-----|---|---|---|
|apachestatus_http|http|http|Checks for apache status pages|
|configjson_http|http|http|Scans for valid `config.json` files|
|dotenv_http|http|http|Scans for valid `.env` files|
|gitconfig_http|http|http|Scans for valid `.git/config` files|
|idxconfig_http|http|http|Scans for `/idx_config` directories with text files|
|laraveltelescope_http|http|http|Scans for open Laravel debuggers|
|phpinfo_http|http|http|Scans for valid `/phpinfo.php` files|
|mysql_open|mysql|open|Connects and checks for default credentials|
|mysql_explore|mysql|explore|Connects and list databases, sizes|
|mongo_open|mongo|open|Connects and checks for open instance|
|mongo_explore|mongo|explore|Connects and list collections, sizes|
|elasticsearch_open|elasticsearch,kibana|open|Connects and checks for open instance|
|elasticsearch_explore|elasticsearch,kibana|explore|Connects and list index, sizes|
|redis_open|redis|open|Connects and checks for open instance|
|kafka_open|kafka}|open|Connects and lists topics|
|couchdb_open|couchdb|open|Connects and list databases, sizes|
|firebase_http|firebase|open|Connects to firebase and checks for `.json` files|@phretor|
|confluence_version|http|http|Scans confluence for vulnerable versions|@HaboubiAnis|
|jira_plugin|http|http|Scans Jira for vulnerable versions|@HaboubiAnis|
|apache_traversal|http|http|Scan servers for Apache LFI|@HaboubiAnis|

### Creating service plugins

Checkout the [l9plugin documentation](https://github.com/LeakIX/l9format/blob/master/l9plugin.md) on how to create your plugins.