# Backend ADR

## Summary


### Issue

We need to choose a back-end programming language that will allow us to transfer data between react-native and firebase

We need a database to store information that is third-party managed to decrease issues after we are finished with project

We need a service to host the server on

### Decision

We are choosing Golang for the back-end

We choose Firebase for the database because it has an admin sdk for go, is easily scalable, and has built in authentification.

We are using heroku to host our API.


## Details


### Assumptions

Golang

  * API endpoints
  
  * Easy unit testing

Firebase

  * Allows for Multiple types of Auth

  * Allows our client to view data and users without having to query a database through firebase console

We want to allow for easy updates and fixes to code.

  * Heroku allows for fast easy developments, deployments, iterations, etc.

### Constraints

We have a constraint that our decision is something to help us complete this project in a semester's time.
We also need something that can be easily transferred when we hand the project over to our clients.


### Positions

We considered the following:

  * Go
  
  * Closure


### Argument

Summary per language:

  * Go: Pros of Go are that is is a general purpose language, open source with lots of documentation, helps with concurrency, statically-typed, and a lower learning curve. Cons of Go the package manager had major problems but has gone under major repairs recently. Prior experience 
  
  * Colsure: Pros of clojure are great program management, general programing, vast amount of packages. Cons of Clojure are there are not as many learning resources available as Go, Difficult syntax, and not a great support community.

### Implications

Backend developer has to learn...

   * How to connect firebase to go through firebase admin sdk
   
   * How to host API on heroku
   
   * How to securely transfer data


