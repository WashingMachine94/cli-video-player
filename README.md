# CLI Video Player
A simple application to watch videos in the terminal.

https://github.com/user-attachments/assets/f64cba4c-7247-4807-9074-f7e04969fe8c

## Installation and Setup

### Prerequisites

- [Go](https://golang.org/dl/) (Make sure Go is installed on your system)

### Building the Application

You can build the application for all supported platforms by running the `build.sh` script. 
The builds will be placed in separate folders based on the target platform.

To run the build script:

#### On Windows:

1. Simply double-click the `build.sh` file.
   
3. The binaries will be created in the `build/` directory under the corresponding platform folders (e.g., `build/linux_arm64/`, `build/windows/`).

 #### On Linux:

1. Ensure that the `build.sh` script is executable. If it's not, you can make it executable by running:

       chmod +x build.sh
   
2. Run the `build.sh` script from the command line:

       ./build.sh

3. The binaries will be created in the `build/` directory under the corresponding platform folders (e.g., `build/linux_arm64/`, `build/windows/`).

### Running the Application on Different Platforms

#### Linux (ARM64 and AMD64)

After building the application, you can set it up for global use so that you can run it from anywhere on your system.

1. Copy the binary to a directory that's included in your `PATH`. Common directories include `/usr/local/bin/` or `/usr/bin/`.

        sudo cp build/linux_arm64/play /usr/local/bin/
   
    For AMD64, you would use:

        sudo cp build/linux_amd64/play /usr/local/bin/
   
3. Ensure the binary has executable permissions:
   
       sudo chmod +x /usr/local/bin/play
4. Now you can run the application globally from anywhere by typing:

       play video.mp4
       play "other video.mp4"
   

#### Windows

After building the application, you can set it up for global use by adding the directory containing the `play.exe` file to your system's environment variables:

1. Build the application using `build.sh`.

2. Add the directory containing the `play.exe` file to your system's environment variables:
- Open the Start Menu, search for "Environment Variables," and select "Edit the system environment variables."
- In the "System Properties" window, click on "Environment Variables."
- Under "System variables," find and select the `Path` variable, then click "Edit."
- Click "New" and add the full path to the directory containing `play.exe`.
- Click "OK" to save the changes.

3. Open a new command prompt window and run the application by typing:

       play video.mp4
       play "other video.mp4"
