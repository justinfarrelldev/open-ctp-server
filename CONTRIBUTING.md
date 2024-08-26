# Contributing

All skill levels are welcome to join us on this epic project! We are always open to contributions. We are also open to contributors who only are available once-in-a-while (I am also only available sometimes)!

# Guidelines

**All contributors must be in the Call to Power Discord!**

This is because there are dedicated sections within the Discord for discussions about development. There is also a Trello board linked in the Discord which we are using for our development activities, Kanban-style.

## Why not Jira?

Most people hate Jira, myself included. In this case, I do not anticipate a large team of developers and there is no Scrum methodology being followed (and no need for extra integrations), so Trello should more than suffice. There is also a Trello webhook for Discord, which is very useful.

## Trello Etiquette

You can: 
- Add comments (please comment your branch name on the ticket in case someone else needs to pick it up for some reason)
- Assign yourself to tickets (so long as someone else isn't already assigned)
- Move tickets between columns as progress is updated (see column meanings below)

You should ask before: 
- Unassigning someone else from a ticket (or taking over their ticket)
- Archiving a ticket
- Modifying the Acceptance Criteria for a ticket
- Modifying the Description of a ticket
- Creating a new ticket
    - We should focus our effort into a few key places to avoid stale tickets
    - New tickets **MUST** have a thorough Description and Acceptance Criteria section

You should **not**:
- Unassign someone from a ticket without consulting them first
- Archive a ticket without bringing it up
- Archive a Trello list for **ANY** reason
- Modifying the Acceptance Criteria for a ticket without asking first
- Modifying the Description of a ticket without asking first
- Share any links to the Trello board outside of the Discord
    - All server developers should be in the Discord for collaboration purposes - it's basically our form of Slack or Teams

Explanations of the columns: 
- To Do: These tickets have been created but have not been assigned / picked up by anyone
- In Progress: These tickets are currently being worked on (or are partially finished)
- PR Created: These tickets have an open Pull-Request in the GitHub repo
- Help Needed: These tickets are "blocked". They need some assistance to be completed. If you are moving your ticket into this section, please also link any relevant GitHub Issues or Discord #development-questions posts so that assistance can come swiftly
- Done: These tickets are done and are queued to be archived once all of the tickets on the board are finished

## Flow

To change code on this project, you must follow a pull-request workflow. 

That is, you should: 
- Fork this project
- Make your changes on that fork
- Make a pull request on this repo for that fork

Once you make a pull-request, a GitHub Actions pipeline will run (which runs the test suite on your code and checks for formatting issues, among other things). If this fails, you should address the issue on your pull request (otherwise, it will not be accepted). 

You should also make sure to check off the action items on the checklist. These are present as reminders to ease along the review process. 

A contributor will then code-review the changes and will suggest changes to your pull request before eventually merging it in.

## Code Spell Checker

There is a Code Spell Checker in the recommended extensions. This will highlight typos with a blue underline to prevent typos from causing problems.

If a word is legitimate correctly spelled, don't hesitate to add it to the dictionary and commit it. These are the steps to do so: 
- Hover over the word with the blue underline
- Click Quick Fix...
- Click on 'Add: "___" to workspace settings'
- Commit this change with your other changes

## Pull Request Guidelines for Reviewers

### Remember the Person

Make sure to treat the person being reviewed with respect. Remember, we are all here out of love for a common series of games and all skill levels are welcome to contribute. Mistakes are allowed to be made.

### Try To Phrase Suggestions as Questions

By this, I mean that instead of saying: 
```
You need to change this line because it's O(n^3).
```
...say instead: 
```
Could we change this function call? It's O(n^3), which may be a bit too slow for map generation purposes.
```

Another example is instead of: 
```
Make this publicly exported
```

...say this instead: 
```
Could we make this publicly exported? We will likely consume this variable in a future update, so it would be super useful!
```

Just generally try to be kind to each other.

### No Swearing / Profanity

This should go without saying, but then again, we may have a Linux kernel contributor amongst us...

## Pipeline Courtesy

If you have recently had a pull request which failed for some reason while being merged to `main`, please attempt to remedy the pipeline. Your code was already approved, so it must be good! If you really can't fix the pipeline by yourself, reach out to Ninjaboy on Discord. 

## Use Gofmt

Please use gofmt (Go's built-in formatter) to format your project.

If you are setting up VS Code for the first time, simply install the recommended Go extension for this repo then bring up the command palette and search for "Format Document". Select Format Document and then select "Go" from the formatter list.

To test that you have gofmt set up correctly, add a few lines in a file and then save. The extra lines should disappear.

## Use Gopls

Please use Gopls within your editor to give yourself syntax highlighting and warnings.

## Do Not Install Other Dependencies Without Discussing It First In Discord

This server should stay as small of a size as possible. Not providing extra dependencies also makes understanding the repo a lot easier.
