This project is minimal and lightweight mobile-first application for scrum teams that wants to
play poker during planning session to find the story point of a task.
So There is options like Coffee, 1, 2, 3, 5, 8, 13, 21, 34.
Users vote and after they vote a admin of session click on reveal button that each vote shown. After talk admin can click on
reset to clean the votes.
There is rooms for each teams that each person can create, users can see all rooms and request to join that owner must accept
to join. Login to application with their phone number and also they can update their name if it's first time to login (signup).
It's must work with websocket and show that in room who is online, also if any user refresh the page it's session must be keep with token
and can reconnect. Any change in room like reveal votes or clean must handle real time.
It's must be minimal and implement with golang and sqlite database.
Some times the owner of page doesn't exists to he can access to some one to be admin of room.
Each room has a name and url to it's session and an avatar that keep in sqlite. Owner can change them.
Admin also can delete users from room.
It's must be mobile first application and light weight and modern design and easy to use in UX. it's design must be dark
and modern and pretty. but again simple.
Don't take it too hard this is simple application and write it with less code.
NOTE THAT THIS APP MUST BE WELL DESIGN IN CODE NOT WRITE ALL FUNCTIONS IN main.go, yes it's minimal and light weight
bust must be well struct and use domain driven design.

Use fiber framework.