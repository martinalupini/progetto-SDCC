## Leader Election Algorithms Implementation

## Starting the simulation

To simulate the execution of a leader election algorithm follow the next steps:

1. Create an EC2 instance.
2. Open one terminal in the directory where you have saved the instance and connect to the EC2 instance:
```
ssh -i <filename>.pem ec2-user@<public-ip>
```
3. Install github on EC2 and clone this repository:
```
sudo yum update
sudo yum install git -y
git clone https://github.com/martinalupini/progetto-SDCC
```
4. Install Docker and Docker Compose by running:
```
./install_docker.sh
```
> [!NOTE]
> After running the script the instance will reboot. After waiting a few seconds, connect again to the istance by typing from your terminal the command at point 1.

Now the environment is ready to start the simulation

## Simulation

At first you need to choose between "Chang-Roberts" and "Bully" by writing in the file **configuration.txt**. If no algorithm is selected the default algorithm is Bully.

In the directory **Scripts/** you can find several scripts to simulate the leader election algorithm.
- To start the simulation type from the directory progetto-SDCC/ (**this step is mandatory**):
```
./Scripts/start.sh
```

Now I suggest to open a new terminal, connect to the istance and run the following scripts.

- After starting the composition there will be a leader elected. I suggest to stop all the containers by typing:
```
./Scripts/stop_all.sh
```
- To see more in detail how the algorithm works you can start one container at a time by typing:
```
./Scripts/start_node.sh <container name>
```
- If you want to stop one container individually (for example the leader to see how the nodes will elect a new leader) you can run:
```
./Scripts/stop_node.sh <container name>
```



