package main

// helpText follows legacy fpm's manual (fpm src/Manual.cs), shortened.
const helpText = `NAME:
    fpm - Flashpoint Component Manager

USAGE:
    fpm [options] <command> [<arguments>...]

COMMANDS:
    list [available|downloaded|updates] [verbose]
        Lists all components: * downloaded and up-to-date, ! downloaded and
        outdated. The filters show only components that are not downloaded,
        downloaded, or outdated; verbose adds the component titles.
    info <component>
        Shows details of a component and whether it is downloaded.
    download [component|category ...]
        Downloads the components (default: all) and their dependencies.
    remove <component|category ...>
        Removes the components.
    update [component|category ...]
        Updates the components (default: all outdated ones) and downloads new
        dependencies and missing core components.
    path [folder]
        Shows or sets the Flashpoint folder in fpm.cfg.
    source [url]
        Shows or sets the component list URL in fpm.cfg.
    verify [component ...]
        Checks that every file of the (default: all downloaded) components exists.
    recover
        Finishes or undoes an interrupted update. This also happens
        automatically before every download, remove or update.
    version
        Shows the version of fpm.

OPTIONS:
    -y, --yes           Do not ask "Is this OK?" (also the default when input
                        is not a terminal).
    --path <folder>     Use this Flashpoint folder instead of the one in fpm.cfg.
    --source <url>      Use this component list URL instead of the one in fpm.cfg.
    -v, --verbose       Show debug messages on stderr.

    Only results go to stdout; messages go to stderr and Logs\fpm.log.

EXIT CODES:
    0 success, 1 error, 2 usage error, 3 some components were skipped,
    4 cancelled.`
