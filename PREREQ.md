# Prerequisites to run the Cadence server locally

You need to install prerequisites which are necessary to run Cadence.   
You need `git`, `docker-compose` and `go` compiler. If you're missing any of them, follow the next session for your OS for step-by-step guidance:

## MacOS

### Install Homebrew (if you don't have it yet):
1. Open **Terminal**
2. Run the following command to install Homebrew:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### Install Git:
Once Homebrew is installed, run the following command to install Git:

```bash
brew install git
````
After installation, verify that Git was installed correctly by checking its version:
```bash
git --version
````
You should see something like git version 2.x.x.

### Install Docker and docker-compose:
1. Go to the official Docker website: [download Docker Desktop for Mac](https://docs.docker.com/desktop/setup/install/mac-install/).
2. Install Docker Desktop:
   - Open the .dmg file you just downloaded.
   - Drag the Docker icon to the Applications folder to install it.
3. Start Docker Desktop:
   - Open the Applications folder and click on the Docker app to launch it.
   - Docker may prompt you to enter your macOS password to complete the installation.
4. Verify Docker Installation:
   - After Docker Desktop starts, you should see the Docker icon in the macOS menu bar.
   - To verify the installation, open a terminal and run:
      ```bash
      docker --version
      ```
5. Verify Docker Compose Installation: Docker Compose is bundled with Docker Desktop, so it should already be installed. To verify, run:
    ```bash
    docker-compose --version
    ````
    This will display the installed version of Docker Compose.

### Install Go:

To install Go compiler (also known as golang), do the run in Terminal:
```bash
brew install go
````

After installation, verify that Go compiler was installed correctly by checking its version:
```go version
```
You should see something like go version go1.xx.x.

## Linux

We assume you're using Ubuntu distribution. If you're using other, you most probably need to change `apt-get` to the package manager used in the distro.

### Install Git:

(Most probably it is already installed if you're using Ubuntu)
1. Open **Terminal**
2. Run the following command to install `git`:
```bash
apt-get install git
````
After installation, verify that Git was installed correctly by checking its version:
```bash
git --version
````
You should see something like git version 2.x.x.

### Install Docker and docker-compose:
1. Run the following in your terminal:
   ```bash
   sudo apt-get install docker-compose
   ````
2. Add current user to docker group:
   ```bash
   sudo usermod -a -G docker $USER
   ````
3. Run the following to apply the above change without re-logging:
   ```bash
   newgrp docker
   ````
4. Verify Docker Installation: Docker Compose is bundled with Docker Desktop, so it should already be installed. To verify, run:
    ```bash
    docker run hello-world
    ````
   This should write some greeting message indicating docker installation is OK and you have enough permissions
5. Verify Docker Compose Installation:
    ```bash
    docker-compose --version
    ````
   This will display the installed version of Docker Compose.

### Install Go:

To install Go compiler, run the following:
```bash
sudo apt-get install golang-go
````

## Windows

### Install Docker and docker-compose:

1. Go to the official Docker website: [download Docker Desktop for Windows](https://docs.docker.com/desktop/setup/install/windows-install/)
2. Install Docker Desktop by executing the file you just downloaded
3. Restart Windows
4. Finish installation of docker by launching Docker Desktop icon
    Proceed installation wizard with Standard options
5. Wait for docker desktop to start (this can take a couple of minutes)
6. Verify Docker Installation:
   - To verify the installation, open a PowerShell and run:
      ```bash
      docker --version
      ```
7. Verify Docker Compose Installation: Docker Compose is bundled with Docker Desktop, so it should already be installed. To verify, run:
    ```bash
    docker-compose --version
    ````
   This will display the installed version of Docker Compose.

### Install Go:

1. Go to the [official Go lang website](https://go.dev/doc/install)
2. Download the latest version, run installer and follow the installation wizard
3. You need to close and run PowerShell again to be able to run `go`
4. Verify that Go compiler was installed correctly by checking its version:
   ```bash
   go version
   ````
   You should see something like go version go1.xx.x.

### Install Git:
Installing Git on Windows is not trivial, though you can always do it yourself.  
In order to save time for the workshop we recommend downloading the repository as a .zip file:
![downloading repo as Zip](https://github.com/user-attachments/assets/0a46d624-b9b6-468c-b3a9-e7e1f3a120a4)
