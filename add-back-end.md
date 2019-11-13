# React Native ADR

## Summary


### Issue

We need to choose a front-end programming language suitable for mobile applications that works well with a backend language like Go.  

Also a language and library that leads to an interactive UI that is inviting and loads quickly.


### Decision

We are choosing ReactNative JS library for the front-end.


### Status

Decided. We are open to new alternatives as they arise.


## Details


### Assumptions

The front-end applications are typical:

  * Typical users and interactions

  * Typical browsers and systems

  * Typical developments and deployments

The front-end applications is likely to evolve quickly:

  * We want to ensure fast easy developments, deployments, iterations, etc.

### Constraints

We have a constraint that our decision is something to help us complete this project in a semester's time.
We also need something that can be easily transferred when we hand the project over to our clients.


### Positions

We considered these libraries of JS:

  * React
  
  * ReactNative
  
  * Angular


### Argument

Summary per language:

  * ReactNative: plenty of freedom, uses a virtual DOM, uses JSX so everything is in one place, only requires JavaScript knowledge and 
  uses a native UI that is fast. Cons are that you have to take care of updates and migrations from other libraries, only one way binding.
  
  * Angular: stable, has two way binding so both UI element and model state change. Cons are that it has less flexibility, 
  requires developer to learn its own syntax and just a web app that can be slow.

### Implications

Front-end developers will need to learn ReactNative. This is likely an easy learning curve if the developer's primary experience is using JavaScript.  Both frontend developers on this project have internship experience with React, so transitioning to ReactNative should not be too different.


