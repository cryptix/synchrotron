# Synchrotron

aims to help with project decentralization and help working offline.

# TODO / Project Scope

* import a list of git repositories (upstreams)
* set a target mirror system. currently planned:
  * ssb
  * ipfs
* worker to poll for updates
* (maybe) webhook to ping back
* (maybe) sign in with github to import starred repos
* (maybe) api parity so that github.com can be replaced with 127.0.0.1

# Rational

Sometimes uplink network speed can be really slow. In these cases fetching all of the dependencies for a project can take more time than you have.
Other kind of network failiures can also leave the upstream unreachable no matter how good local connectivity is.
The end result is the same: access is hindered and work has to pause.

Even though git is a decentral system where you can work with what you already have cloned locally, you still need upstream access for updates.

Modenm decentralization systems like IPFS and SSB take the _offline first_ aspect seriously. Nodes in these networks have all their data available locally so no uplink connectivity is required once data is fetched and in sync.

The pruposed solution to above problem is to make a bot that mirrors a specified list of repositories onto such a system and have users of this but pull updates from this bot.

# Fork of QOR example application

see [qor-example](https://github.com/qor/qor-example).

## License

Released under the MIT License.

