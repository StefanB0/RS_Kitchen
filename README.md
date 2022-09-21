# RS_Kitchen

Laboratory written by Boicu Stefan FAF 203

Prerequisites for building the services: Docker installed and running on the machine. https://docs.docker.com/get-started/

To build the kitchen image Run 

`$docker build -t kitchen-image .` if you are on windows

or  

`$sudo docker build -t kitchen-image .` if you believe in Linux supremacy

Then you need to build the image for the dinning-hall. Refer to README.md in that project.

After you build both images, to start all the services run

`$docker compose up` on windows

or

`$sudo docker compose up` on linux
