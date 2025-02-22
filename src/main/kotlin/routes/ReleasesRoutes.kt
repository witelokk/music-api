package com.witelokk.routes

import com.witelokk.models.FailureResponse
import com.witelokk.models.Release
import com.witelokk.models.ShortArtists
import com.witelokk.models.Songs
import com.witelokk.tables.*
import io.ktor.http.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import models.ShortArtist
import org.jetbrains.exposed.sql.select
import org.jetbrains.exposed.sql.transactions.transaction
import java.util.*

fun Route.releasesRoutes() {
    authenticate("auth-jwt") {
        route("/releases") {
            get("/{id}") {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val releaseId = try {
                    UUID.fromString(call.parameters["id"])
                } catch (e: IllegalArgumentException) {
                    return@get call.respond(
                        HttpStatusCode.NotFound, FailureResponse("release_not_found", "Release not found")
                    )
                }

                val release = getReleaseWithArtist(userId, releaseId)
                    ?: return@get call.respond(
                        HttpStatusCode.NotFound, FailureResponse("release_not_found", "Release not found")
                    )

                call.respond(release)
            }
        }
    }
}

fun getReleaseWithArtist(userId: UUID, releaseId: UUID): Release? {
    return transaction {
        val release = Releases.select { Releases.id eq releaseId }.singleOrNull() ?: return@transaction null

        val artists = ReleaseArtists.leftJoin(Artists)
            .select { ReleaseArtists.releaseId eq releaseId }
            .map { artistRow ->
                ShortArtist(
                    id = artistRow[Artists.id].toString(),
                    name = artistRow[Artists.name],
                    avatarUrl = artistRow[Artists.avatarUrl]
                )
            }

        val songs = ReleaseSongs
            .select { ReleaseSongs.releaseId eq releaseId }
            .map { getSongWithArtistsAndFavorite(it[ReleaseSongs.songId], userId)!! }

        Release(
            id = release[Releases.id].toString(),
            name = release[Releases.name],
            coverUrl = release[Releases.coverUrl],
            type = release[Releases.type].name,
            releasedAt = release[Releases.releasedAt].toString(),
            songs = Songs(
                count = songs.size,
                songs = songs
            ),
            artists = ShortArtists(
                count = artists.size,
                artists = artists
            )
        )
    }
}
