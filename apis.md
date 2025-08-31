Create these apis:
1. User login with email and if it's new user it's response with false ProfileCompleted. 
2. User fill it's name and it's avatar (store avatars in a data/upload directory and store it's path to database to serve.)
3. Create greeting apis and it's most be constant for each user in 1 to 15 minutes:
    use this for welcome message:
    ```go
        if hour >= 5 && hour < 12 {
            greetingOptions = []string{
                "Morning, %s! Ready to plan?",
                "Hey %s! Fresh start?",
                "Hello, %s! What's brewing?",
                "Rise and shine, %s!",
                "Morning, %s! Let's roll!",
                "Hey %s! Coffee and code?",
                "Hello, %s! New day ahead!",
            }
        } else if hour >= 12 && hour < 17 {
            greetingOptions = []string{
                "Hey %s! Afternoon vibes!",
                "Hello, %s! Still crushing it?",
                "Afternoon, %s! What's next?",
                "Hey %s! Midday momentum?",
                "Hi %s! Keep pushing!",
                "Hello, %s! In the zone?",
                "Hey %s! Let's build!",
            }
        } else if hour >= 17 && hour < 22 {
            greetingOptions = []string{
                "Evening, %s! Still at it?",
                "Hey %s! Wrapping up?",
                "Hello, %s! Night shift?",
                "Evening, %s! What's cooking?",
                "Hey %s! Second wind?",
                "Hi %s! Evening plans?",
                "Hello, %s! Still going strong?",
            }
        } else {
            greetingOptions = []string{
                "Night owl, %s!",
                "Hey %s! Burning midnight oil?",
                "Late night, %s! Still up?",
                "Hello, %s! Night coding?",
                "Hey %s! Can't sleep?",
                "Hi %s! Quiet hours!",
                "Hello, %s! Stars are out!",
    }
    }
    ``` 
   also send qoutes that are store in qoutes.md file in random
4. Use can send request to join room and owner of room can accept or reject it. So create api to send request to join.
5. Create api to owner see all requests and accept or reject it.
6. Create api to user see it's pending requests.
7. Create api to user see all rooms with two array one is my room and another is other room.
8. User can create room by add name and a picture of it's avatar.
9. Owner can make admin a user and revoke it.
10. Owner can make some one to be admin
11. User vote in room and just admin and reset votes or reveal them.
12. Websocket that live show actions in room like votes and reset and other thing.
13. Api to show all users in room and who are now online in current room (not in enitre aplication in room, must be a flat to show who is online).
14. Only if admin click on reveal the api send the votes.
15. Api to setting the room just by admin, in that he can change the room and it's avatar or remove members or even delete the room.
16. Api to show the profile of user and login api and can edit it's api.