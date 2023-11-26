
# Trust and Trustability

This is part of a project to formulate a practical Promise Theory model of trust for our Internet and machine enabled age. It is not related to blockchain or so-called trustless technologies, and is not specifically based on cryptographic techniques. Rather it addresses trustworthiness as an assessment of reliability in keeping specific promises and trust as a tendency to monitor or oversee these processes.

The code presented here is free to use, but it is to be thought of as a proof of concept, since much of it is tailored to the specific use cases for which data could be obtained. Although the code has been made to be general and reusable, it isn't at the level where is can simply be plugged into any software with a beneficial result. Much of what has been learned from developing the code indicates a simple approach that agrees quite well with the state classification model used in the CFEngine configuration management software (an early AI/ML based system assessment engine), yet even after 25 years of experience we don't know how to use it yet. So, I reckon we're still at the learning stage.

The goal here is to investigate how we might use trust as a guiding potential in human-information systems.  For the semantic elaboration, the specific example of users interacting with Wikipedia to read and to write contributions is used for concreteness.  They should be viewed in parallel with the papers and documents available at http://markburgess.org/trustproject.html

The code consists of a directory of stubs and analysis tools, that make of the Go package library called TT, which is built from the earlier Semantic Spacetime library code and new methods developed in the test cases here.

If you're not a Go(lang) afficionado (I'm certainly not religious about it either), think of it as a better and faster Python. You'll need to set up some basics to run the code:

For a detailed discussion, you might like to read this post: https://medium.com/@mark-burgess-oslo-mb/universal-data-analytics-as-semantic-spacetime-16fbfad4f5de

## Running the code:

My working environment is GNU/Linux, where everything is simple. Setting up the working environment for all the parts is a little bit of work (more steps than are desirable), but it should be smooth.

1. Install git client packages on your computer.
2. Go to: https://golang.org/dl/ to download a package.
3. Some file management to create a working directory and link it to environment variables:

```
  $ mkdir -p ~/go/bin
  $ mkdir -p ~/go/src
  $ cd ~/somedirectory
  $ git clone https://github.com/markburgess/Trustability.git
  $ ln -s ~/somedirectory/Trustability/pkg/TT ~/go/src/TT
```

4. It's useful to put this in your ~/.bashrc file
```
export PATH=$PATH:/usr/local/go/binexport GOPATH=~/go
```
Don’t forget to restart your shell or command window after editing this or do a
```
  $ source ~/.bashrc
```
5. You can get fetch the drivers for using the graph database code and the 
```
  $ go get github.com/arangodb/go-driver
  $ go get github.com/sashabaranov/go-openai
```
6. Finally (thank goodness) you'll need the database itself to run some of the examples. Go to
```
  https://www.arangodb.com/download-major/
```