# 🍽️ FoodBridge 

**Connect. Donate. Feed.**  
A lightweight platform to bridge the gap between surplus food from restaurants and those who need it, powered by volunteers.

---

## 🚀 Overview

FoodBridge allows:
- **Restaurants** to list surplus food with pickup details.
- **NGOs** to browse and accept available donations.
- **Volunteers** to pick up food from restaurants and deliver to NGOs.
- **Admins** to manage users, donations, and reset passwords.

Built as a fully functional MVP for hackathons – simple, clean, and ready to demo.

---

## ✨ Features

- 🔐 Role‑based registration & login (Restaurant, NGO, Volunteer, Admin)
- 📋 Restaurant dashboard – create & track donations
- 🤝 NGO dashboard – accept available donations
- 🚚 Volunteer dashboard – accept pickups & update delivery status
- 📞 Contact details (phone numbers) visible to all parties after acceptance
- 🔄 Auto‑refresh every 15 seconds on all dashboards
- 🛡️ Admin panel – view all users/donations, edit/delete, view password hashes, reset passwords
- 🎨 Modern, responsive UI with Font Awesome icons
- 📡 REST API built with Go (Gin framework)
- 🗄️ MongoDB for data storage
- 🔑 JWT‑based authentication

---

## 🛠️ Tech Stack

| Frontend        | Backend         | Database  | Auth  |
|-----------------|-----------------|-----------|-------|
| HTML, CSS, JS   | Go (Gin)        | MongoDB   | JWT   |
| Font Awesome    | REST API        |           | bcrypt|

---
🔄 Workflow
1.Restaurant lists surplus food (name, quantity, pickup                time/location).
2.Donation appears as available.
3.NGO browses available donations and accepts one → status becomes     accepted.
4.Accepted donations are visible to volunteers.
5.Volunteer accepts a pickup → status becomes assigned.
6.Volunteer picks up food → marks as picked_up.
7.Volunteer delivers → marks as delivered.

At each step, all parties see relevant contact numbers for coordination.

