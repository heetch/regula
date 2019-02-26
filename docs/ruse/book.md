# The Book of RUSE
## Introduction
Once upon a time, in a company far, far away, their was a product manager called Blanche.  Blanche works with a team of 7 software engineers, named Heureux, Somnolent, Éternuement, Grincheuse, Timide, Dopé, and Dave.  Together they craft an app that allows the elderly to arrange rides on carriages from the crypto-mines to the city and back.    Blanche trusts and respects the software engineers, and believes in their insistence that she must wait for long periods between releaseses of their software whilst they finely craft every detail.  However, many times Blanche becomes frustrated because she doesn't have the ability to fine tune her software running in production.  Conditions are always changing, and she'd like to have the app reflect it.  Here are some example cases Blanche has noticed in her data stream:

- On Mondays there are very few elderly folks looking for a ride to the city. 
- Wood cutters are not allowed to ride, but there are often tired wood cutters ask about the service. 
- On Thursdays pensions are paid in the city and the carriages will all be fully occupied.
- In the autumn some of the elderly folk like to use the service to travel to orchards and return with apples.
- Poisoning of carriage drivers is a reoccurring issue.  It only seems to happen in the autumn. 

Blanche would like to try out some ideas to help improve the service and reduce the number of carriage drivers being poisoned.  Sadly the release cycle of her app is simply too long to allow these tweaks and experiments to be carried out in a reasonable way.  One day however, when she was moaning about this to Somnolent, Grincheuse overheard, and she interjected to say:

> Oh do shut up moaning Blanche.  Just use Regula, you idiot.

Now, Blanche was used to Grincheuse's little outbursts, so she thought nothing of her rough words, but she enquired. 

> Oh dear Grincheuse, a regular what?

To which Gincheuse replied:

> Nah, you bloomin' nincompoop, not "Regular", "Regula", it's a bleedin' rules engine, init. 

Blanche looked confused:

> But what is a rules engine?

At this point Somnolent, who had woken from all the commotion, stepped in:

> A rules engine... 

.. he paused to stretch and yawn ... 

> ... is a way to put some specific logic outside of a computer
> program, and have that program respect it.  The program gives the
> rules engine the relevant facts, and asks the rules engine to make a
> decision.  The way that decision is made is up to the person
> controlling the rules, not the program, so the rules can change at
> any time.

Blanche stopped and pondered this for a while.  She casually brushed a bluebird from her hair, where it had landed and become entangled, and jumped a little and she narrowly avoided kicking the rabbits that had gathered around her ankles. Then she said:

> Oh, that sounds exactly what I need, but Somnolent, Grincheuse how
> could I possibly define the logic for the rules, I am not a software
> engineer like you.  I am too tall, to go down the software mines for
> starters!

Grincheuse turned back to Blanche with a steely look:

> Are you still blathering on?  Look it's easy, even for an numskull
> like you.  There's a very simple, high level language called RUSE.

Blanche was beginning regret asking, but at that moment Timide, who
had been waiting quietly at Blanche's side for the last minute or so
cleared his throat and handed Blanche a book.  It was called "The Book
of RUSE".  Blanche opened it eagerly and speed-read the introduction. 

> Hang on?

.. she said...

> This is the conversation we've just been having.

Éternuement farted. 

## Of signatures and software engineers

Blanche found a quiet spot in the corner of the office, and sat down on a big sofa.  She opened the book and began to read the second chapter.  After reading a few paragraphs, she noticed that the seven software engineers were pushing aside the woodland animals that usually accompanied her, and were pulling up stools around her.  She had the creeping, uncomfortable realisation that this was becoming a meeting.  She thought she'd better give the meeting some direction:

> It says here that before I can write rules I have to agree a rule-set signature with the software engineers.

Grincheuse didn't attempt to disguise her smirk as she responded:

> Only if you want the app to actually do something when you provided rules. 

Luckily Heureux took a more conciliatory tone:

> Yes, that's quite right Blanche. The app has to know when to consult
> Regula for a decision.  It also need to know what information its
> supposed to give to Regula to support making that decision, and what
> type of information it's supposed to give back.

Blanche thought a little and then asked:

> If the app has to know that up front, how can this help me add dynamic behaviour to the app?

Heureux smiled kindly and answered:

> This is a very important point.  Regula and RUSE aren't there to make the app do new things, we still determine what the app *can* do, but the rules determine *what* it should do in a given situation.  They're like an especially flexible settings panel for the app. 

Blanche nodded sagely, then continued.

> So what does it mean that you need to know the *type* of information
> the rule set gives back?

