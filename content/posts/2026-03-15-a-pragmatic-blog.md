---
title: "A Pragmatic Blog"
date: 2026-03-15
tags: [blog, go, pragmatism, git-crypt, cli]
description: "Building this blog and the wonders of pragmatism"
---
_Can we talk about my blog? I'm dying to talk about my blog._

![iasip-charlie-the-mail](/images/iasip.jpg)

## Background rant

I decided to finally get around to building out a space where I can write out my thoughts.
I've been writing code for a while now (10+ years), a lot of your mindset changes as you get more experienced
as you keep learning about new technologies and the ecosystems pushes forwards you get caught up in the complexity of it.

_"Yea to send this account setup email lets create an email service so we can isolate all email-related code from the other parts of the codebase,
oh shit what if our request fails in transit to the email service, lets put a message queue in between! We might only have 100 users now, but think about the future!"_

I might be beating a dead horse here, but I really don't think it can be said enough. Complexity is alluring, you think you are doing yourself a favor building
a system that can work at scale, before you have scale because you won't have to come back to it in the future where you might have a thousand other things going.
So you build for scale day 1, with the good intentions of not having to worry about it later. I'm not saying you don't need a message queue or an email service, but
98% of software don't. It might look nice on your resume but a few years down the road and you are going to regret that you didn't just use a simple solution.

Okay, I got a bit sidetracked there. What I'm trying to say is that there is an absolutely wonderful world of **dead simple** solutions to common problems, where people will reach
for "industry standards" where you might not even need it.

---

## Building the thing

When I set out to build this blog I had a few features I knew I wanted:

* Public blog posts _(shocker)_
* Private/Draft posts
* Anonymous comments
* Email subscribers
* Admin tools/Moderation

When I sat down to actually begin working on the thing, I hadn't really thought deeply about how I wanted each of these features
to function. It's very typical for how I usually build things, it's more of a creative process for me and I enjoy uncovering the right
way to build the thing as I'm building it. The original plan was pretty webdev basic/industry standard.
Oh yea we need private posts? Lets spin up postgres, add a 'published' boolean column on the posts, easy?
Writing posts on the go? Of course! We'll just add the standard jwt web login so I can login, write my posts and moderate comments.

I've built things this way for so long now, why do I need a whole web login? Who is going to log in to this thing? It's just me, am I really going to be
writing a blog post from my phone? _(I'm not)_ The more I thought about it the more the whole thing bugged me, so much complexity for something that should just be simple.
In the end I went a whole different direction, completely public repo, no web login, no admin panel, no postgres _(still got a db though)_ and I still managed to get every feature I wanted
but with way less code, less complexity and some hidden bonus advantages.

---

### Posts are just files

How many posts am I really going to write? More than likely we are talking sub thousands, completely no reason to reach for a database for that. Instead posts can just be markdown files. Way simpler to reason about, less complexity around handling them **and** its comes with the added benefit of having all of my content backed up **free** of charge via my git provider. Githubs built in editor is perfectly fine for making changes or adjustments to posts too, if I ever did want to actually edit a page on the go, I can simply do it straight on github and continuous deployment takes care of publishing my changes.

---

### Private/Draft posts are invisible

Okay my posts are now _just_ files, but then how do we keep draft or journal entries private? Since the repo is on github
I could just private it, but I dislike that for several reasons:

* A single _'security'_ point of failure, any exposure and it all fails.
* Lost options for contributions. Someone might spot a mistake in the code or a typo in a post.
* Potentially tied to whatever git service you decide to use and if they have a private option.
* Harder to deploy since you now need to be authenticated.

_> How can we keep the repo public and have all of our hidden content as files?_

Encryption! I did a bit of searching and found [git-crypt](https://github.com/AGWA/git-crypt) which uses gits native clean/smudge filters to run a command before commit/checkout. This automatically encrypts any posts placed in the `content/private/` folder in our `.git` blobs which means I don't have to manage it all and all posts are available in my editor, private or not but encrypted
on any remote git server. Pulling on a fresh machine just requires me to set the encryption key in my git settings and run `git-crypt unlock`, all our posts are now decrypted. Hosted we simply run without the key set and only our public posts are available. It just **works**.

---

### Admin tooling is dead simple

I wanted people to be able to interact with the blog and be notified whenever something new was posted, for this I _did_ opt for a database but went with sqlite to stay minimal and simple. The scale of a personal blog is easily handled by sqlite and having the database as just another file feels right. All of our posts e.g. The main content is still all backed up in our git provider, losing the database would not lose us the main parts of the blog.

This is where it gets tempting to reach for an admin dashboard of some kind and I would usually think about building this into the site but thats a web auth layer again and a whole slew of complexity in login pages, sessions/cookies and so on. Instead I opted to do a dead simple CLI with a simple API Key auth _'layer'_. Set the key in your hosted instance and in your local .env along with the url of the instance and you have a convenient way to quickly do simple moderation and keeping track of subscribers:

```
# .env

BLOG_URL=https://thobiasn.dev
ADMIN_API_KEY=thats_pretty_neat
```

```
❯ ./blog
usage: blog <command>

Commands:
  serve                          start HTTP server
  new post <title>               create a new post (in content/private/)
  new project <name>             create a new project
  publish <slug>                 move post from private to public
  dash                           admin dashboard
  comments                       list recent comments
  comments delete <id>           delete a comment
  comments toggle <id>           toggle comment visibility
  subscribers                    subscriber stats
```

A few handy commands for writing locally, and a handy dashboard for a quick glance.

```
Blog Dashboard
──────────────
Public posts:  2
Private posts: 1
Comments:      0
Subscribers:   1
```

---

### Just right

Make the whole thing in go so we get a single binary with a low footprint. The same binary I use locally to write the posts is the one deployed on the remote server. It's just the exact level of complexity needed, which is piss all for a simple private blog and its all in under 2000 lines of go code.

It **feels** just right.
