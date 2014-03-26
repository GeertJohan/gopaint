
var state = 0;
var canvas = {};
var timestamps = [];
var ws = new WebSocket("ws://"+document.location.host+"/socket");

ws.onopen = function() {
	console.log("websocket opened");
};

ws.onmessage = function (msg) { 
	console.dir(msg);
	switch(state) {
		case 0:
			state = 1;
			canvas = jQuery.parseJSON(msg.data);
			
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
						style: "border: 1px dotted #303030;  width: "+sizeX+"%; height: "+sizeY+"%; position: absolute; left: "+posX+"%; top: "+posY+"%;",
					});
					$('#body').append(pixelDiv);
				}
			}

			// update pixels
			for (layerName in canvas.Layers) {
				if (canvas.Layers.hasOwnProperty(layerName)) {
					var layer = canvas.Layers[layerName]
					layer.Pixels.forEach(function(pixel) {
						if(!timestamps.hasOwnProperty(pixel.Position) || pixel.Timestamp > timestamps[pixel.Position]) {
							timestamps[pixel.Position] = pixel.Timestamp;
							console.log('pixel '+pixel.Position+' is #'+pixel.Color);
							$('#pixel_'+pixel.Position).css('background-color', '#'+pixel.Color);
							$('#pixel_'+pixel.Position).attr('title', layer.Name);
						} else {
							console.log('old pixel');
						}
					});
				}
			}
			break;
		case 1:
			update = jQuery.parseJSON(msg.data);

			// var layer = {};
			// if(canvas.Layers.hasOwnProperty(update.LayerName)) {
			// 	layer = canvas.Layers[update.LayerName];
			// } else {
			// 	layer = {
			// 		Name: update.LayerName,
			// 		Pixels: [],
			// 	};
			// 	canvas.Layers[update.LayerName] = layer
			// }

			// if(!canvas.Layer.Pixels.hasOwnProperty(update.Position)) {
			// 	canvas.Layer.Pixels[update.Position] = {};
			// }
			var color = "";
			if(update.PixelState) {
				color = '#'+update.PixelColor;
			} else {
				color = "white";
			}
			$('#pixel_'+update.PixelPosition).css('background-color', color);
			$('#pixel_'+update.PixelPosition).attr('title', update.LayerName);

			break;
	}
};

ws.onclose = function() {
	alert("websocket connection with server is closed...");
	$('#body').html("<h3>Connection lost. Please refresh this page.</h3>");
};