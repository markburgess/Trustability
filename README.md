
# Trust and Trustability

See also the separate reusable library TnT which extracts the main methods for reuse.

This is part of a project to formulate a practical Promise Theory model of trust for our Internet and machine enabled age. It is not related to blockchain or so-called trustless technologies, and is not specifically based on cryptographic techniques. Rather it addresses trustworthiness as an assessment of reliability in keeping specific promises and trust as a tendency to monitor or oversee these processes.

The code presented here is mainly for research purposes. It is free to use, but it is to be thought of as a proof of concept, since much of it is tailored to the specific use cases for which data could be obtained. Although the code has been made to be general and reusable, it isn't at the level where is can simply be plugged into any software with a beneficial result. Much of what has been learned from developing the code indicates a simple approach that agrees quite well with the state classification model used in the CFEngine configuration management software (an early AI/ML based system assessment engine), yet even after 25 years of experience we don't know how to use it yet. So, I reckon we're still at the learning stage.

## Documentation of the detailed experimental process

The goal here is to investigate how we might use trust as a guiding potential in human-information systems.  For the semantic elaboration, the specific example of users interacting with Wikipedia to read and to write contributions is used for concreteness.  They should be viewed in parallel with the papers and documents available at http://markburgess.org/trustproject.html

This can be found on the summary page http://markburgess.org/trustproject.html, which refers to methodology and data sets. All data sets can be generated at any time, though of course the data changes gradually over weeks and months.

1. [Literature summary](https://www.researchgate.net/publication/369185404_Background_Notes_on_The_Literature_of_Trust_Bridging_The_Perspectives_of_Socio-Economics_and_Technology)
2. [Notes on Incorporating Operational Trust Design into Automation (Companion to code examples)](https://www.researchgate.net/publication/371292051_Notes_on_Incorporating_Operational_Trust_Design_into_Automation_Companion_to_code_examples)
3. [Promise Theory model of trust and trustworthiness (complete draft for peer review)](https://www.researchgate.net/publication/370303770_Trust_and_Trustability_-v01_An_idealized_operational_theory_of_economic_attentiveness)
4. [Narrative Attentiveness Notes on Analysis and Assessment of Semantic Trustworthiness](https://www.researchgate.net/publication/372110659_Narrative_Attentiveness_Notes_on_Analysis_and_Assessment_of_Semantic_Trustworthiness_v01_Companion_notes_to_the_ngram_code_examples)
5. [Trust assessment examples (Notes accompanying examples and code](https://www.researchgate.net/publication/372589267_Trust_assessment_examples_Notes_accompanying_examples_and_code)
6. [Group collaboration of users on Wikipedia A data study of trust dynamics (Notes accompanying examples and code)](https://www.researchgate.net/publication/373237181_6_Group_collaboration_of_users_on_Wikipedia_A_data_study_of_trust_dynamics_Notes_accompanying_examples_and_code)
7. [Studying attack resilience in human-machine processes using (mis)trust(worthiness) as a predictor (The security angle)](https://www.researchgate.net/publication/374030005_7_Studying_attack_resilience_in_human-machine_processes_using_mistrustworthiness_as_a_predictor_The_security_angle)
8. [Code consolidation](https://www.researchgate.net/publication/375074930_8_Consolidate_code_from_Semantic_Spacetime_Model_Put_data_into_queryable_database)

9. [Machine Learning Model for simplifying trust allocation: the role of heuristics](https://www.researchgate.net/publication/375927262_Can_we_learn_when_to_trust_New_research_on_relating_trust_machine_learning_and_heuristics)
10. [Summary of the project: A Trust User Guide](https://www.researchgate.net/publication/376271829_A_Trust_User_Guide)

## Code and data

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