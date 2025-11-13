Go migrate   ( use should istall it in order to run commands)  (everyone)

INSTALLATION


go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

-tags (Type of DB we are using. for our project 'sqlite' )

Migration  How it works

we keep all the files in the folder so  the go migrate can keep track of our file (latest-Newest)  

You can think of it as a Git for databases 

this is achieved automatically when you create from the command line the file from migrate 
                 
Command :  migrate create -ext sql -dir db/migrations -seq create_posts_table

-ext  flag   :  Extension for the file  example it add (.sql) to create_posts_table

-seq flag : it creates the file with a numerical value so it can keep track of the changes example : 000001_create_users_table.up.sql

if you dont specify the -seq flag it will create the file and it will keep track from the timestamp 

example : 20241003153045_create_posts_table.up.sql

- The Migrate Create command   it actually creates 2 files 

example : 

000001_create_users_table.up.sql
000001_create_users_table.down.sql

.up file ---->  it is used for the changes that we want to make 

example : i want to add a collum   in the table posts

.down file ----> it is used  for reverting the change we made  so we should write the code that brings the table to the previous state 



How to Run  .up file 
                (example)                               (folderpath example)
migrate -path db/migrations -database "sqlite://C:/Users/Morie/Desktop/Zone-01/social_network/social_network.db" up

what happends when you run this command


-The migrate tool looks in the migrations folder for migration files.
It checks which migrations have already been applied to your database.


-It applies any new .up.sql migration files (in order) to your SQLite database.

-The database schema is updated according to the SQL in each .up.sql file.

-A record of applied migrations is stored in a special table in your database, so migrations are not re-applied.

note that there is an extra table in the Database that keeps track


How to run .down files  (revert changes)

migrate -path db/migrations -database "sqlite://C:/Users/Morie/Desktop/Zone-01/social_network/social_network.db" down 1

-This command will undo the last applied migration by running the corresponding .down.sql file.

-The number 1 means "revert one migration." You can change this number to revert more migrations at once.

-The database schema is reverted according to the SQL in each .down.sql file.

-The migration tracking table is updated to reflect the reverted migrations.


Note  SOS : you can't brake the order  
for example if i have made a change in user table and after at the post table 
If i wanted to revert the user table i have to revert the post table also  (like versions in git)  

so in this case is better to create a new migration that changes the user table 


-keep in mind in some Databases like sqlite if you want to revert a changes 

for example i added a collum in a table

the .down file should look like this  

-- Remove 'title' column from posts (SQLite workaround)
CREATE TABLE posts_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO posts_new (id, user_id, content, created_at)
SELECT id, user_id, content, created_at FROM posts;

DROP TABLE posts;
ALTER TABLE posts_new RENAME TO posts;



what it does it creates a new table  without the collum  and adds the data to the new table then Drops the old table  and renames the new one with the olds name 


