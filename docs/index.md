# Regula

## Overview

The "Regula" project addresses the problem caused by the accumulation of simple but recurrently changing logic due to business requirements. This introduces strains on the development cycle by reducing the velocity. In concrete terms, developers can find themselves always chasing the the business rather than focusing on core logic which defines the business.

A solution to his kind of problems is a "Business Rules-Engine" or "Production Rule System". It solves it by providing a way to extract logic by creating "rules" that dictates a result under precise conditions.

Such rules are ordered and composed of a condition and an action - simplistically, it's a bunch of ordered if-then statements. [Learn more about Rule-Engines](https://martinfowler.com/bliki/RulesEngine.html).

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
  // Codebase using Regula will receive those updates live
  if city == "france-paris" AND driver-status == "gold" then 3.0
  if city == "france-paris" then 2.0
  if city == "italy-milan" then 1.5
  // catch all clause, acts as a default
  if true then 1.0
```

### How does it help being productive

So once developers and stakeholders have agreed on what simple logic will be extracted, it is possible to split the ownership in two:
// By extracting simple logic into the Rules-Engine, it is possible to split the ownership in two:
- "when and how it's being used": developers own the context 
- "how is it computed ": product people own the behaviour

Schema: before after 

Regula takes a strongly opiniated approach in implementing this solution in order to provide strong predictability and reliability. Those are detailed in the "design principles" section below.


### What Regula provides

Regula is a solution available to backend developers, mobile developers as well as on the front-end. 

Currently, those are the supported environments: 

- Backend
  - Go
  - Elixir
- Mobile
  - through Go-mobile

## Terminology

A **rule** is a condition that takes parameters and if matched, returns a **result** of a certain type. 

It can be illustrated as following in pseudo-code:

```
  # string city and string driver-status are parameters
  if city == "france-paris" AND driver-status == "gold" then 3.0
```

A **ruleset** is an ordered list of **rules** that return results of the same type. It is named which allows to identify it. Rules are evaluated from top to bottom until one matches. 

This is a **ruleset** named `/marketplace/radius`
```
  if city == "france-paris" AND driver-status == "gold" then 3.0
  if city == "france-paris" then 2.0
  if city == "italy-milan" then 1.5
  if true then 1.0
```

A **signature** is the combination of a ruleset name, its parameters and their types and finally the result type. Once defined, it's set into stone and cannot be changed (to do so, create another ruleset).

It can be illustrated as following (still in pseudo-code):

```
float64 /marketplace/radius(string city, string driver-status)
```


A **version** is an identifier that enables to refers to a specific version of a ruleset. Each time a ruleset has at least one of its rules updated, a new **version** is created to identify it.

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
  if city == "france-paris" then 4.0
  if city == "italy-milan" then 3.0
  if true then 1.0
```
