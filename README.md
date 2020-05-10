# Movie Recommend Telegram Bot

A Telegram bot which picks a movie from my watchlist.

Implemented with Golang!

## Live Demo
You can try the bot on [this](https://t.me/CKMovieRecommendBot) link.

Just drop a message `/recommend` and check the recommended movie!

## Used Services
- [JotForm](https://api.jotform.com): The data store for watchlist, stored in a spreadsheet.
It keeps and returns  the title, year and director information for movies.
The bot randomly selects one movie, nothing smartie.
- [OMDB](https://www.omdbapi.com): After picking a random movie, the bot fetches other details
(such as movie poster, duration, plot etc) from the OMDB api.

## Motivation
Instead of sitting in front of a list of movies and trying to select one,
I decided to implement a bot to pick one randomly.

Also, it was a great oppurtunity to get hands-on experience with golang.

## License
This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
