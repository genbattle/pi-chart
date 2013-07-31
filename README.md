pi-chart
========

A basic server to display basic 2D Raster/Vector content on the Raspberry Pi via HTTP. This is intended to be used to display graphs and charts on a TV as a basic kiosk/heads-up sort of display.

To current TODO list/some ideas:
 - Get HTTP server listening for POST requests.
 - Allow Images to be POSTed directly.
 - Allow Image URLs to be POSTed.
 - Create a JSON or XML-based format for sending raw data to be graphed:
	- Bar
	- Line
	- Pie
	- Points (Scatter)
	- Others?
	- Images
 - Allow multiple content pieces to be put together in a single POST and displayed in different parts of the screen.
 - Create an update-only mode where layout and display is set, but datasets change dynamically.
