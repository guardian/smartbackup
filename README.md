# smartbackup

Backup app for Postgres databases using NetApp backed storage

## Why? What does it do?

Database backups are a pain. Traditionally you run pg_dump (or equivalent) to make a huge SQL file that can be replayed,
and compress it.  It takes ages, needs loads of disk, and takes even longer to replay.

In our Multimedia system the database is over 600G and this becomes so problematic as to make the backup almost useless.

Fortunately, Netapp has a much better solution - snapshotting.

We can tell the Netapp to preserve the current storage blocks as a "snapshot" - no data is copied, only locked, so this 
is a fairly quick operation.

Of course, the problem is that the database is still writing.  Under normal circumstances for this to work (i.e. for us to
be able to actually restart a b0rked system from the snapshot) the app in question must be "quiesced" i.e. prevented from
writing during the snapshot window.

Fortunately, Postgres is clever about this; it does not need to be prevented from writing but it _does_ need to create a
"consistent restore point".  When an instance is brought up from the snapshot it discards data from after the "restore point"
as potentially corrupt and attempts to replay as much as possible from any source of write-ahead logs available
(see the Postgres documentation for many, many, more details on this).

So, our backup process becomes:

```
    ----------------------      ------------------      ---------------------------------------
   | Set db restore point | -> | Perform snapshot | -> | Tell database that backup is complete |
    ----------------------      ------------------      ---------------------------------------
```

This is a simple app to manage this process

## SnapCenter

Netapp offer a product which can do this process for a whole range of databases and present a nice UI along with it.
Individual databases are supported by plugins which sit in a Java agent that needs installing on the database box.

Unfortunately we have had a lot of problems:
 - the logs are difficult to read
 - the plugin installation and running has trouble with selinux
 - extra firewall holes are required for the agent
 - you can't use Ansible, Salt or any other config management platform to deploy easily, you're meant to do it through their UI
 - Postgres is a "community supported" plugin which uses shell script, embedded in Perl, called from Java, called from a server....
 - In a nutshell, we could not get it to work reliably.
 
But then, Netapp released OnTap 9.6 which brings with it a REST API.  And so a simpler way was born....


## How does it work?

1. Provide a config file - see `example-config.yaml` in the repo root for details on how this must be formatted and what
information is required
2. The config file gives the login details for Netapp appliances/SVMs and Postgres databases and a list of "targets",
i.e. what needs to be backed up from the above lists
3. Run `./smartbackup`.
4. Each target is run through in turn, a restore point set, a snapshot made and the backup completed.
5. - If a restore point fails, then that backup fails and we move on to the next. 
   - If the snapshot fails then the backup is still completed.
   - If the backup complete operation fails then it is just logged.
   
## Where do i install it?

Wherever you want!  It accesses both NetApp and Postgres over network connections so it can run anywhere that connectivity
is availabile.  However, you need to include (limited) plaintext credentials in the configuration files so it should be
hosted somewhere secure in your environment.  There is no reason NOT to run it locally on your database server and in
some cases this might indeed be the best place to do it since you can use the local "postgres" admin privilege to avoid
extra user accounts.  On the other hand, a secure little place or a Kubernetes cluster would work equally well.

It's assumed that the app will be run from a task scheduler like cron

## How do I build it?

You'll need Go 1.11 or higher installed to build - https://golang.org/dl/

With that installed, simply clone this repository and run `go build` in the root.  This will give you a binary
called `smartbackup` in the same directory, built for your local environment.
For (really easy!) information about cross-compiling see https://golangcookbook.com/chapters/running/cross-compiling/

### TL;DR

GOOS=linux GOARCH=arm go build

See https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63 for supported values