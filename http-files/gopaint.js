
var state = 0;
var ws = new WebSocket("ws://"+document.location.host+"/socket"+document.location.search);

ws.onopen = function() {
	console.log("websocket opened");
};

$('#postInfo').html('post to: http://'+document.location.host+'/draw<br/>more info: github.com/GeertJohan/gopaint')

ws.onmessage = function (msg) { 
	console.dir(msg);
	switch(state) {
		case 0:
			state = 1;
			var canvas = jQuery.parseJSON(msg.data);
			var timestamps = [];

			if(canvas.SinglelayerMode) {
				$('#viewAllLayers').html('<a href="/">View all layers</a>');
			}

			$('.showWidth').html(canvas.Width+'');
			$('.showHeight').html(canvas.Height+'');

			// calculate sizes in percentage
			var sizeX = 100/canvas.Width;
			var sizeY = 100/canvas.Height;

			// create canvas
			for(x=0; x<canvas.Width; x++) {
				for(y=0; y<canvas.Height; y++) {
					var n = y*canvas.Width + x;
					var posX = x*sizeX;
					var posY = y*sizeY;
					var pixelDiv = $("<div>", {
						id: "pixel_"+n,
						style: "width: "+sizeX+"%; height: "+sizeY+"%; left: "+posX+"%; top: "+posY+"%;",
						class: "pixel",
					});
					$('#canvas').append(pixelDiv);
				}
			}

			// update pixels
			for (layerName in canvas.Layers) {
				if (canvas.Layers.hasOwnProperty(layerName)) {
					var layer = canvas.Layers[layerName]
					layer.Pixels.forEach(function(pixel) {
						if(!timestamps.hasOwnProperty(pixel.Position) || pixel.Timestamp > timestamps[pixel.Position]) {
							timestamps[pixel.Position] = pixel.Timestamp;
							$('#pixel_'+pixel.Position).css('background-color', '#'+pixel.Color);
							$('#pixel_'+pixel.Position).attr('title', layer.Name);
						}
					});
				}
			}
			break;
		case 1:
			update = jQuery.parseJSON(msg.data);

			$('#pixel_'+update.PixelPosition).css('background-color', '#'+update.PixelColor);
			$('#pixel_'+update.PixelPosition).attr('title', update.LayerName);

			break;
	}
};

ws.onclose = function() {
	alert("websocket connection with server is closed...");
	$('#canvas').html("<h3>Connection lost. Please refresh this page.</h3>");
};