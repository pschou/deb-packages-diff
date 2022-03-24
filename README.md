# Debian Packages Diff

This tool detects the differences between two Debian Packages.gz files and prints out the deltas.

## Examples

Using two Packages.gz files
```bash
$ ./deb-package-diff -new NEW_Packages.gz -old OLD_Packages.gz -showAdded
```

Using one repo directory and one file
```bash
$ ./deb-package-diff -new output/ -old OLD_Packages.gz -showAdded -output filelist.txt
```

Using just a new file, this gives you a full list
```bash
$ ./deb-package-diff -new NEW_Packages.gz -old "" -showAdded
```

## Usage
```bash
$ ./deb-package-diff -h
Debian Package Diff, Version: 0.1.20220323.2200

Usage: ./deb-package-diff [options...]

  -new string
        The newer Packages.gz file or repodir/ dir for comparison (default "NEW_Packages.gz")
  -old string
        The older Packages.gz file or repodir/ dir for comparison (default "OLD_Packages.gz")
  -output string
        Output file for comparison result (default "-")
  -repo string
        Repo path to use in file list (default "dists/Debian11.2/main/binary-amd64")
  -showAdded
        Display packages only in the new list
  -showCommon
        Display packages in both the new and old lists
  -showRemoved
        Display packages only in the old list
```
