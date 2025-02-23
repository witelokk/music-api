package com.witelokk.music.tables

import org.jetbrains.exposed.sql.Table

object Favorites : Table("favorites") {
    val userId = uuid("user_id").references(Users.id)
    val songId = uuid("song_id").references(Songs.id)

    override val primaryKey = PrimaryKey(userId, songId)
}