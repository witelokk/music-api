package com.witelokk

import com.witelokk.tables.*
import org.jetbrains.exposed.sql.Database
import org.jetbrains.exposed.sql.SchemaUtils
import org.jetbrains.exposed.sql.transactions.transaction

fun connectToDatabase(url: String, user: String, password: String = "") {
    Database.connect(
        url = "jdbc:$url",
        driver = "org.postgresql.Driver",
        user = user,
        password = password,
    )
    println("Database connected successfully")

    transaction {
        SchemaUtils.create(Users)
        SchemaUtils.create(Artists)
        SchemaUtils.create(Followers)
        SchemaUtils.create(Songs)
        SchemaUtils.create(Favorites)
        SchemaUtils.create(SongArtists)
    }
}
