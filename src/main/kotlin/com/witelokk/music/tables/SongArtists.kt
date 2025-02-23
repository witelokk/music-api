package com.witelokk.music.tables

import org.jetbrains.exposed.sql.Table

object SongArtists : Table("song_artists") {
    val songId = uuid("song_id").references(Songs.id)
    val artistId = uuid("artist_id").references(Artists.id)

    override val primaryKey = PrimaryKey(songId, artistId)
}