# Vectortile view component

## Development server

Run `ng serve` for a dev server. Navigate to `http://localhost:4200/`. The application will automatically reload if you change any of the source files.

## Build

Run `ng build` to build the project. The build artifacts will be stored in the `dist/` directory.

## Embedding the webcomponent

Build the project and copy the style and javascript files into the application embedding the download component

Embed the webcomponent in a third party web application

- load styles and javascript in your html

  ```html
   <link rel="stylesheet" type="text/css" href="vectortile-view-component/styles.css">
    <script type="text/javascript" src="vectortile-view-component/main.js"></script> 
    <script type="text/javascript" src="vectortile-view-component/polyfills.js"></script> 
    <script type="text/javascript" src="vectortile-view-component/runtime.js"></script> 

    <app-vectortile-view style="width: 800px; height: 600px;"
    tile-url="https://api.pdok.nl/lv/bag/ogc/v0_1/tiles/NetherlandsRDNewQuad" zoom=12 center-x=5.3896944
    center-y=52.1562499>
  </app-vectortile-view>

see index.html for other samples.
  
  ```


  ```
