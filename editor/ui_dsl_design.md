# UI Editor

A UI editor written in Golang using cimgui-go (AllenDang).

+ Sprites Editor; which allows you to edit "sprites" and define named regions for each sprite, along with their positions and sizes.
+ Font Editor; which just lists "ASCII -> glyph" mappings and allows you to import TTF files and generate bitmap fonts.
+ Variable Editor; which allows you to define variables that can be used in the UI for dynamic elements, like temperature, humidity, 

## Scripting Language for UI Design

We need a specialized scripting language to define the UI connectivity matrix and the drawing instructions for each page. The syntax should be simple and intuitive, allowing to easily create and modify UI layouts without needing to write complex code. But mainly it allows us to tweak the menu while the APP is running, which is a huge boost for development speed.

- line; mode, start, end, color
- rectangle; mode, top_left, bottom_right, color
- circle; mode, center, radius, color
- text; mode, font-name, size, position, text, color
- sprite; effect, image, position, size
- component; name, position
- optionals; prompt-id, position, options...

## Script

# ==========================================
# UI CONNECTIVITY MATRIX
# ==========================================
matrix
  .          PageB_L2   .
  PageA_L2   PageB_L2   PageA_L2
  PageA_L1   PageB_L1   PageC_L1   PageA_L1
  PageA      PageB      PageC      PageA
end

# ==========================================
# CONSTANT DEFINITIONS
# ==========================================
screen
  WIDTH  800
  HEIGHT 600
end

colors
  # Colors
  COLOR_RED    #FF0000
  COLOR_GREEN  #00FF00
  COLOR_BLUE   #0000FF
  COLOR_WHITE  #FFFFFF
  COLOR_BLACK  #000000
end

positions
  # Positions
  POS_TOP_LEFT     0 0
  POS_TOP_RIGHT    WIDTH 0
  POS_BOTTOM_LEFT  0 HEIGHT
  POS_BOTTOM_RIGHT WIDTH HEIGHT
  POS_CENTER       WIDTH/2 HEIGHT/2
end

# ==========================================
# FONT DEFINITIONS
# ==========================================
fonts
  # Id size file_path
  Roboto16 16 assets/Roboto-Regular.ttf
  Roman24 24 assets/Roman-Regular.ttf
end

# ==========================================
# SPRITE ATLAS DEFINITIONS
# ==========================================
sprites assets/player.png
  # Id x y width height
  player_idle 0 0 32 32
  player_run 32 0 32 32
  cloud 0 32 64 32
end

sprites assets/buttons.png
  # Id x y width height
  buttonA 0 0 200 50
  buttonA_hover 0 50 200 50
end

# ==========================================
# COMPONENTS
# ==========================================
component nice.button
    # line   [mode]  [start]  [end]    [color]
    line     stroke  0 0      200 50   COLOR_BLACK
    
    # rect   [mode]  [top_left] [bottom_right] [color]
    rect     fill     0 0      200 50       COLOR_GREEN
end

# Mode: fill, stroke
# Effect: normal, invert, greyscale, mirror

# optionals
# Render-time interactive selector command.
# Every render pass, user is asked to pick components from the option list.
# If user cancels or selects none, nothing is rendered for this command.
#
# Syntax:
# optionals [prompt-id] [x] [y]
#   option [component-name]
#   option [component-name]
# end

# =========================================================================================
# PAGE DRAWING INSTRUCTIONS
# Instructions are processed in order, so you can draw a background first and then overlay 
# it with sprites &components
# =========================================================================================
page PageA
  # line   [mode]  [start]  [end]    [color]
  line     stroke  0 0      100 50   COLOR_RED

  # rect   [mode]  [top_left] [bottom_right] [color]
  rect     fill     10 10      200 150       COLOR_GREEN

  # circle [mode]  [center] [radius] [color]
  circle   stroke   400 300  50       COLOR_BLUE

  # text   [font]     [position] [color]     [text]
  text     Roboto24   15 40      COLOR_WHITE "{Temp}{Celsius}"

  # sprite [effect]  [name]       [position] [size]
  sprite   greyscale player_idle  120 45     32 64

  # component [name] [position]
  component nice.button 300 400

  optionals weather_picker 50 420
    option nice.button
    option warning.card
  end
end

page PageB
  line fill  0 0      800 600 COLOR_BLACK
  rect fill  0 0      800 600 COLOR_WHITE
  circle fill 400 300  100 COLOR_RED
  text Roman24 20 20 COLOR_WHITE "MAIN MENU"
  sprite invert buttonA 100 100  200 50
end

page PageC
    line fill 0 0 800 600 COLOR_BLACK
    rect frame 0 0 800 600 COLOR_WHITE
    circle fill 400 300 100 COLOR_GREEN
    text Roman24 20 20 COLOR_WHITE "SETTINGS"
end

