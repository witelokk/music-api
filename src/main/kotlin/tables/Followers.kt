package com.witelokk.tables

import org.jetbrains.exposed.sql.Table

object Followers: Table("followers") {
    val userId = uuid("user_id").references(Users.id)
    val artistId = uuid("artist_id").references(Artists.id)

    override val primaryKey = PrimaryKey(userId, artistId)
}