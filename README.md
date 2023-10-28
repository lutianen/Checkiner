# Checkiner

This is a checkin script for some useful web sites.

**You MUST change config file including the filename, email and password.**

**Please DON'T expose the config file to others, or your account may be stolen.**

```bash
# config/example
example@example.com
ThisIsAPassword
```

## Build

```bash
mkdir bin
gob -o bin/checkiner src/checkiner.go src/main.go src/utils.go 
```

## [autostart](https://wiki.archlinuxcn.org/wiki/KDE#%E8%87%AA%E5%90%AF%E5%8A%A8)

```bash
# /home/tianen/.config/autostart/checkiner.desktop
YOUR_PROJECT_DIR = /home/tianen/go/src/Checkiner # NOTE YOU MUST CHANGE THIS DIR
[Desktop Entry]
Exec=$YOUR_PROJECT_DIR/bin/checkiner -w THY@CUTECLOUD -p $YOUR_PROJECT_DIR/config/THY@$YOUR_PROJECT_DIR/config/CUTECLOUD
Icon=
Name=checkiner
Path=
Terminal=False
Type=Application
```
