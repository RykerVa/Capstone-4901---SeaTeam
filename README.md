# Capstone-4901---SeaTeam

Go mod file needed to run projects on your machine.
  - In a new directory run 'go mod init NewGoProject' replacing NewGoProject with your project name.
  - From here you can add your Go files to this module.


To view tracing:

- Download Docker
  - Using recommended settings
- Ensure packages are installed and mod file has the neccesary requirements 
- Start zipkin ex) 'docker run -d -p 9411:9411 openzipkin/zipkin'
- In a web browser go to 'localhost:9411' using the example above

