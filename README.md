## Leader Election Algorithms Implementation

## Starting the simulation

To simulate the execution of a leader election algorithm follow the next steps:

1. Create an EC2 instance
2. Open one terminal and connect to the EC2 instance 
```
ssh -i <filename>.pem ec2-user@<public-ip>
```
3. Install github on EC2 and clone this repository
```
sudo yum install git -y
git clone https://github.com/martinalupini/progetto-SDCC
```
4. Choose between "Chang-Roberts" and "Bully" by writing in the file **configuration.txt**. If no algorithm is selected the default algorithm is Bully.
5. Initialize the istance and start the containers by running
```
chmod +x configure_instance.sh
./configure_instance.sh
```
6. To stop containers individually run
```
./stop <container_name>
```

