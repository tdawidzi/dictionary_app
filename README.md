# GraphQL API - PL-ENG dictionary simulation
This is a Dictionary application that provides a GraphQL API for managing words and their translations. It allows you to add words, translations, examples, and query them in different languages using GraphQL.

## Technologies Used

- **Go (Golang)**: The backend of the application is built using Go.
- **GORM**: Object Relational Mapping (ORM) used for database interactions.
- **PostgreSQL**: Database used to store words, translations, and examples.
- **GraphQL**: API layer for querying and mutating data.

## Database schema
The application uses the following models:

- **Word**: A word in a specific language.
- **Translation**: A translation between two words in different languages.
- **Example**: An example sentence using a word.

![Database schema](https://github.com/tdawidzi/dictionary_app/blob/dev/Dictionary_database.svg)

## How to use
### Step 1: Clone the repository

Clone the repository to your local machine:

```bash
git clone https://github.com/tdawidzi/dictionary_app.git
cd dictionary_app
```

### Step 2: Create the .env file
The .env file provides sensitive information about database connection.
Change name of .env.example file to .env, and modify environmental variables included in it.
SAVE CHANGES

### Step 3: Build and start API
Run the following command (in main project folder) to build and start the application:
```bash
docker-compose up --build
```
This command will start two containers:
- **WordPostgreSQL**: The database container running on localhost port specified in .env (default: 5432)
- **APP**: The GO app running on localhost:8080

Wait for log in terminal: 
```
app-1       | Successfully created tables
app-1       | Server listening on: http://localhost:8080/graphql
```
This means, that everything is working correctly.

(DB migrations are performed Automatically)

### Step 4: Access aGraphQL API
Once containers are running correctly you can access GraphQL api at: ```http://localhost:8080/graphql```.
For test you can use any GraphQL Client as Altair or GraphQL Playground

## Possible queries and mutations
### Managing Words
List all words:
```
 query {
   words {
     id
     word
     language
   }
 }
```
Add word:
```
mutation {
  addWord(word: "kot", language: "pl") {
    id
    word
    language
  }
}
```
Modify word:
```
mutation {
  updateWord(oldWord: "caat", newWord: "cat" language: "en"){
    word
  }
}
```
Delete word:
```
mutation {
  deleteWord(word: "elelephant", language: "en")
}
```
### Managing Translations
List all translations for given word:
```
query {
  translationsForWord(word: "kot") {
    word
    language
    }
}
```
Add translation:
(Adding translation is only 'connecting' two words in database - to create translation, both words have to be previously created)
```
mutation {
  addTranslation(wordPl: "kot", wordEn: "cat") {
    id
  }
}
```
Modify translation:
```
mutation{
  updateTranslation(
    oldWordPl: "kot"
    oldWordEn: "dog"
    newWordPl: "kot"
    newWordEn: "cat"
  ){
    id
  }
}
```
Delete translation:
```
mutation {
  deleteTranslation(wordPl: "kot", wordEn: "cat")
}
```
### Managing Examples
List all examples for given word:
```
query {
  examplesForWord(word: "kot") {
    id
    example
  }
}
```
Add example:
```
mutation {
   addExample(word: "kot", language: "pl", example: "Kot śpi na kanapie."){
    id
    example
  }
}
```
Modify example:
```
mutation{
  updateExample(
    id: 1,
    example: "Kot leży na kanapie."
  ){
    id
    example
  }
}
```
Delete example:
```
mutation{
  deleteExample(id: 1)
}
```