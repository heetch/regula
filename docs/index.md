# Regula

## Overview

The "Regula" project addresses the problems caused by the accumulation of simple logic that changes frequently in line with business requirements. This introduces strains on the development cycle by reducing the velocity. In concrete terms, developers can find themselves always chasing the business rather than focusing on core logic which defines the business.

A solution to these kind of problems is a "Business Rules engine" or "Production Rule System". This provides a way to extract logic by creating "rules" that dictates a result under precise conditions. 

Such rules are ordered and composed of a condition and an action - simplistically, it's a bunch of ordered if-then statements. [Learn more about Rule Engines](https://martinfowler.com/bliki/RulesEngine.html).

What really makes this approach shine is that stakeholders can change the rules themselves, without having to wait for developers to implement it. This enables much faster development cycles.

Example: Let's say there's a ride-sharing company that needs to adjust the radius in which it can match drivers and passengers. This changes a lot, depending on various real life events:

```
Given I want to know "What radius should I match driver and passenger with" in my code 
And I don't want to hardcode the logic
When I query the ruleset named "/ride-marketplace/radius"
And that I submit "product", "driver-status" and "city"
Then I get a radius value based on rules that have been defined by product owners

---

radius = 
  // Product people can update these rules on their own
  // Codebases using Regula will receive those updates live
  if city == "france-paris" AND driver-status == "gold" then 3.0
  if city == "france-paris" then 2.0
  if city == "italy-milan" then 1.5
  // catch all clause, acts as a default
  if true then 1.0
```

### How does it help productivity

Once developers and stakeholders have agreed on what simple logic will be extracted, it is possible to split the ownership in two:

- **"when and how it's being used"**: developers own the context. Developers are responsible on how that logic is being used inside the codebase.
- **"how is it computed "**: stakeholders own the behaviour. Stakeholders can update the logic themselves with total autonomy. 


The following diagram highlight this split, with a **ruleset** that computes a radius. 

![Before/After schema](./before_after.png)


| Before                                                            | After |
|-------------------------------------------------------------------|-------|
| 1⃣ Developers and stakeholders agree on a given solution   | 1⃣ Same  |
| 2⃣ Stakeholders write specifications                        | 2⃣ Stakeholders and developers agree on a **signature**: which parameters the ruleset should take as well as its return type: **Developers create the ruleset on Regula.**
| 3⃣  Developers implement the solution                      | 3⃣ Using Regula client, developpers evaluate the ruleset in their code by passing the correct parameters. **(Much shorter!)**|
| 4⃣ Stakeholders update the specifications when the situation changes  | 4⃣ **Stakeholders directly update the ruleset when needed** |
| 5⃣ Developers update the code accordingly                   | ✨ **Developers are not involved at all** |

The key point to remember here is that once a **ruleset** has been created, stakeholders can never break the contract of taking specific paremeter and returning a specific type. They can only change how the result is computed. It's that specific constraint that allows to split ownership. 

Regula takes an opinionated approach in implementing this solution in order to provide strong predictability and reliability. Those are detailed in the [design principles](#design-principles) section below.


### What Regula provides

Regula is a solution available to backend and mobile developers as well as on the front-end. 

Currently, these are the supported environments: 

- Go
- Elixir
- Mobile Apps, through Go-mobile

## Terminology

A **rule** is a condition that takes parameters and if matched, returns a **result** of a certain type. 

It can be illustrated as follows in pseudo-code:

```
  # string city and string driver-status are parameters
  if city == "france-paris" AND driver-status == "gold" then 3.0
```

A **ruleset** is an ordered list of **rules** which return results of the same type. It is named which allows to identify it. Rules are evaluated from top to bottom until one matches. 

This is a **ruleset** named `/marketplace/radius`
```
  if city == "france-paris" AND driver-status == "gold" then 3.0
  if city == "france-paris" then 2.0
  if city == "italy-milan" then 1.5
  if true then 1.0
```

A **signature** is the combination of a ruleset name, its parameters and their types and finally the result type. Once defined, it's set in stone and cannot be changed (to do so, create another ruleset).

It can be illustrated as following (still in pseudo-code):

```
float64 /marketplace/radius(string city, string driver-status)
```


A **version** is an identifier that refers to a specific version of a ruleset. Each time a ruleset has at least one of its rules updated, a new **version** is created to identify it.

**Ruleset version: ** `4`
```
  # ruleset: /marketplace/radius
  if city == "france-paris" AND driver-status == "gold" then 3.0
  if city == "france-paris" then 2.0
  if city == "italy-milan" then 1.5
  if true then 1.0
```

Let's update the results on certain rules: 

**Ruleset version: ** `5`
```
  # ruleset: /marketplace/radius
  if city == "france-paris" AND driver-status == "gold" then 6.0
  # a rule can updated in both its conditions and result
  if city == "france-paris" OR city == "france-lyon" then 4.0
  if city == "italy-milan" then 3.0
  if true then 1.0
```

## Creating and updating rulesets

Stakeholders can agree with developers and they will create a ruleset. Once created, stakeholders can edit rulesets with full autonomy, with the guarantee of not breaking the code. Updating a ruleset means, in concrete terms to modify one or many rules, by changing their conditions or their result. 

The decision to use Regula for a given business problem should be made by answering the following questions:

- the behaviour often changes or is supposed to evolve a lot
- the behaviour is relying entirely on business facts 

If the answer is yes for both of them, then creating a ruleset is a good idea. 


In the current release of Regula (`0.5.0`) stakeholders can only create or edit rules through scripts which require some engineering knowledge. Later version will include a proper UI for stakeholders to interact with Regula in a simple manner. 


## Components overview 

> ⚠ TODO: we should discuss what we want to bring in here, this can get quite complex really quick.

![Cinematic](./cinematic.png)

- Regula clients query Regula server
  - they live in the source code of a given application
  - when high performance is required, i.e. ruleset evaluations should be as fast as possible, there are caching mechanism available in some clients. See more about this in the technical documentation.
  
- Regula servers answer requests from the clients. 
- Ruleset storage is where the rulesets and their versions are kept. It is not supposed to be interacted with directly.

## Design principles

Rule engines are delicate beasts. It is important to remember that it's about extracting behaviour from the source code and relocating it outside in a more "accessible" place that is outside of any continuous integration mechanism. 

Regula takes an opinionated approach to lower the risks and make it harder to break from expected behaviours. Regula's team objective is to focus on a curated feature set, putting reliability and predictability at a top priority. 

## Predictability 

Regula needs to be predictable because it splits the ownership of given functions of the software that uses it. Behaviour being owned by whoever edits the rules, it is mandatory that a given ruleset is not able to break the software that uses it. 

On the developers side:

Regula is "typed" and that is why the concept of **signature** exists, illustrated by:

```
float64 /marketplace/radius(string city, string driver-status)
```

By doing so, it allows developers of both typed and dynamic languages to be certain that they always get a value of a given type. Changing the signature implies the creation of a new rule and such a break will logically requires modifying the source code of the application using Regula to reflect the change. 

On the stakeholders side: 

Regula rulesets are "versioned", meaning every update to rulesets are tracked. By doing so, it is possible for stakeholders to observe how the behaviour is evolving. 

Moreover, versioning allows stakeholders to take versions into consideration when designing a product. A given business object could be stored with the version of rulesets it is intended to use throughout its life. For example, if a given promotion coupon had been applied to a product being sold, by storing the version used, it can easily be computed again against that version, allowing to get the same exact result. 

By doing so, it enables stakeholders to know which path was taken to compute a given value, whether it's in the application logs or stored within a business object. 

## Reliability

Regula depends on two facts to be reliable: rulesets storage should be reliable, Regula itself should be reliable. 

Therefore: 

- All data is stored in an etcd store, which **should be distributed properly** (at least three instances in production, ideally on multiple avaibility zones).

- There should be multiple instances of Regula, **which should be distributed properly** (at least two instances in production, ideally on multiple availability zones, so there is no single point of failure).
