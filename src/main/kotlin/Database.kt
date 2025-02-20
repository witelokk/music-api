package com.witelokk

import com.witelokk.tables.Users
import org.jetbrains.exposed.sql.Database
import org.jetbrains.exposed.sql.SchemaUtils
import org.jetbrains.exposed.sql.transactions.transaction

fun connectToDatabase() {
    Database.connect(
        url = "jdbc:postgresql://localhost:5333/",
        driver = "org.postgresql.Driver",
        user = "postgres",
//        password = "postgres"
    )
    println("Database connected successfully")

    transaction {
        SchemaUtils.create(Users)
    }
}
