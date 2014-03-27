## Go Paint!

Go paint is a simple server program that allows anyone to draw pixels by making an HTTP request.

The canvas is displayed in a browser, and is updated in 'real time' using a websocket.

### draw
To draw on the canvas, make an HTTP request to the server with path `/draw`
Parameters are given as `GET` fields in the URL.
 - `x`: the x position of the pixel to edit
 - `y`: the y position of the pixel to edit
 - `color`: the color (6 character hex) to set
 - `layer`: your name, or something that identifies you

This piece of software has been build for fun and should not be used for production.
There's no server-side persistency, please don't be mad when your Mona Lisa disappears..