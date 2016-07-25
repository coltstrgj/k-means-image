# K Means Theme Generator

This project is just a start. It is some code that I threw together in a couple of hours to play with. I may continue to update it (in fact I have plans to) but I might not as well. 

It uses a modified k means cluster analisys to generate a color scheme from an image. 

## Usage

You can run the program with no arguments to see options. 

Basic usage is k-means-image input.png

## Math Basics

### Clustering 

This program uses almost exclusively integer math. I have not had any problems with convergence so far, and could easily be changed, but the RGB values of images are integers, so I stuck with it for now. I treat the R, G, B values as X, Y, Z respectively. This effectively maps every pixel in a 3d space. I then cluster them.

Each point (pixel) keeps track of the cluster it is currently associated with (it's center) and the centers keep track of how many members they have. The cluster centers do not keep track of which members belong to them. The cluster centers also keep a running average of the X,Y,Z coordinate of member points. This is updated every time a point joins or leaves the cluster. This is the part of the math where int's could potentially be a problem. The RGB values are already 16 bit, so the loss in accuracy I get from using int math is pretty much taken care of with this extra int precision.


### Brightening Steps
After convergence I (optionally with -b num) try to brighten the clolrs. The colors that are generated up to this point  will be an average of all of the similarly colored pixels. This means that they are usually very dull colors (grey, browns, whites and blacks). The brightening step fixes this problem by choosing a color that has a larger gap between Max(r, g, b) and min(r, g, b). This is probably not the best way to do this, but I do not know much about colors, so this is the best I could come up with. 

This process is best demonstrated with an example. Let us say that I run cluster analysis and get the following clusters:

1. Members: 30, average RGB: 55, 55, 55
  * 10 members are pixels with color (45, 45, 45)
  * 10 members are pixels with color (55, 55, 55)
  * 10 members are pixels with color (65, 65, 65)
2. Members: 10, average RGB: 200, 200, 225
  * 5 members are pixels with color (200, 200, 200)
  * 5 members are pixels with color (200, 200, 250)

That is to say that I got two very bland colors generated. If I were to increase the number of colors to 4, I would still get 4 bland colors because the number of brightly colored pixels (where R, G, and B values are dissimilar) is smaller than the number of them that is not, so I cannot just increase the number of clusters. Instead, I take each cluster and using only the points in that cluster I run the program again generating more clusters. I could then do this again with these new clusters as many times as the user specifies. I find that 1 or two steps is usually enough. Then I go through each cluster and find the which one is the brightest, even if it has a low number of members, and get rid of the rest. This means that our 'solution' may not contain all points from the image any more 

1. Members: 20, average RGB: 50, 50, 50
  * 10 members are pixels with color (45, 45, 45)
  * 10 members are pixels with color (55, 55, 55)
2. Members: 5, average RGB: 200, 200, 250
  * 5 members are pixels with color (200, 200, 200)

This does not contain all 40 pixels from the original image, but all of them were included at some point or another. 

I will add images later that demonstrate this more clearly. 
