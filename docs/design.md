# Asset Server

- An asset server that serves assets, sprites, fonts, palettes, over TCP to connected clients.

## Functional Design

- asset-server server.config.json
  - The server reads a configuration file, prepares (convert) the assets and then serves the assets to clients over TCP.
- server.config.json contains paths to:
  - fontpak.config.json
  - spritepak.config.json
- fonts can be converted to SDF or coverage bitmaps, and packed into a single binary chunk
- sprites can be packed into a single binary chunk
- palettes can be packed into a single binary chunk
- asset-server sends one continuous binary chunk to the client, which contains all fonts, sprites and palettes in a single binary blob. The client doesn't have to do any de-serialization or parsing, it is a load-in-place binary blob that can be used directly by the client.
- the config also contains a path to a ccova script that can be compiled and the byte-code image can be send to the client as part of the binary blob. The client can then execute the ccova script to initialize the assets and perform any other initialization tasks.


## ESP32

- Boot
- Activate WiFi, connect to Asset Server EP
  - Upon connection, the Asset Server immediately uploads the AssetDb chunk, the Script chunk and the
    state chunk to the ESP32.
- The ESP32 then goes into setup(), initializes any sensors, display, touch and the script
- loop() is then called, which ticks the script and performs any other tasks, during this the ESP32
  cannot receive anything over WiFi but small information packets (UDP?) containing data like:
  - weather state, inside and outside humidity/temperature/pressure, date & time, light state. 
    All of this state is one single binary piece of memory < 1.2 KB that is frequently sent to every
    ESP32 over WiFi. The ESP32 can then use this data to update the display, and perform any other tasks.
