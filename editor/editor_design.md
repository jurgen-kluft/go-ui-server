# UI Editor

A UI editor written in Golang using cimgui-go (AllenDang).

Editor has the following components:

* Script Editor (actively saves the script to disk)
* Sprites List (can add/remove/edit)
   * Upon pressing edit a window appears as a Sprite Editor where we can add/remove sprites and name/move/resize them, we should also be able to zoom in
* Fonts List (can add/remove fonts)
* Window showing the UI, sized and rendered accordingly from the active Script
