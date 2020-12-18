# l9 suite stock plugins

[![GitHub Release](https://img.shields.io/github/v/release/LeakIX/l9plugins)](https://github.com/LeakIX/l9plugins/releases)
[![Follow on Twitter](https://img.shields.io/twitter/follow/leak_ix.svg?logo=twitter)](https://twitter.com/leak_ix)

This repository contains LeakIX maintained plugins implementing the [l9format golang plugin interface](https://github.com/LeakIX/l9format/blob/master/l9plugin.go).
They are currently used by [l9explore](https://github.com/LeakIX/l9explore) but could be implemented by Go security tool.

## Service plugins

|Plugin|Protocols|Stage|Description|
|------|-----|---|---|
|mysql_open|mysql|open|Connects and checks for default credentials|
|mysql_explore|mysql|explore|Connects and list databases, sizes|
|mongo_open|mongo|open|Connects and checks for open instance|
|elasticsearch_open|elasticsearch,kibana|open|Connects and checks for open instance|
|redis_open|redis|open|Connects and checks for open instance|

### Creating service plugins

Checkout the [l9plugin documentation](https://github.com/LeakIX/l9format/blob/master/l9plugin.md) on how to create your plugins.

## List of web plugins

### WIP