# Purpose
For beginning coders, working in the terminal can be daunting and unfamiliar. IDEs (Integrated Development Environments) help to ease some of this burden by providing a more intuitive and user-friendly method for doing development work, and they are widely used by many professional software developers. This document aims to help guide less experienced developers with instructions on how to set up Katabole in VSCode and Cursor. 

# Prerequisites
To begin, you'll need to install either [VSCode](https://code.visualstudio.com/) or [Cursor](https://cursor.com). Cursor is a fork of VSCode, with more AI features baked in. Feel free to use whichever one you prefer.

# Quick Start
After installing your preferred IDE, open the Katabole project inside the project by going to File > Open, then select the location of where you cloned your Katabole fork on your local computer.

Inside of your IDE, go to the menu bar at the top and select "Terminal > New Terminal". Run the following commands (also listed in the [README](README.md)):

```bash
go install github.com/katabole/katabole@latest
katabole gen --import-path github.com/myuser/myapp --title-name MyApp
cd myapp
task setup

task dev
```

# Tips
It is also strongly recommended that you install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go) inside of your IDE. This provides IntelliSense features as well support for testing and debugging.
