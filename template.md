# 🌟 Framework Insights

Stay at the forefront of the development landscape with **Framework Insights**, a meticulously curated ranking of the most popular frameworks. The popularity score for each framework is determined by evaluating a set of key metrics that reflect its community engagement, adoption, and overall usage.

🔄 **Contribute:** To add or update tracked frameworks, simply modify the [`projects.txt`](projects.txt) file.

📜 This project is licensed under the [MIT License](LICENSE).

---

## 📋 Trending Frameworks

{{.Table}}

---

**⏰ Last Updated:** {{.LastUpdated}}  
_This page is updated automatically every 24 hours to ensure the latest framework rankings are displayed._

---

## 📂 JSON Data Availability

Looking for programmatic access to this data? The framework rankings are also available in JSON format in the [`frameworks.json`](frameworks.json) file.

---

## 📊 Score Calculation

The popularity score of each framework is calculated by considering various factors such as Stars, Forks, Watchers, Subscribers, and Issues. Each of these metrics has a weighted contribution to the final score, giving a more comprehensive view of the framework's standing in the developer community.

- **Stars (40%)**: The number of stars a framework has on GitHub, indicating its popularity.
- **Forks (25%)**: The number of forks, which represents how often the framework is being used as a base for other projects.
- **Watchers (20%)**: The number of watchers, reflecting interest in the repository for updates.
- **Subscribers (10%)**: The number of users subscribed to the repository for notifications on new changes.
- **Issues (5%)**: The number of open issues, with a higher number of issues increasing the overall score.

The framework's score is calculated using a weighted average of these metrics. Each metric is multiplied by its respective weight, and the total score is normalized to fall between 0 and 1.

Example formula:

```
Score = (Stars * 0.4) + (Forks * 0.25) + (Watchers * 0.2) + (Subscribers * 0.1) + (Issues * 0.05)
```

---

## 🤝 Contribution

We welcome contributions from the community! To contribute:

1. Open the [`projects.txt`](projects.txt) file.
2. Add a new line with the GitHub repository URL of the framework. Each framework should have its own line.  
   Example:

   ```
   https://github.com/facebook/react
   https://github.com/angular/angular
   ```

3. Save the file, commit your changes, and submit a pull request.

---

## ⚖️ License

This project is open source and available under the [MIT License](LICENSE). Feel free to use, share, and contribute!
