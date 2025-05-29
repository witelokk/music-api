package com.witelokk.music.routes

import com.witelokk.music.models.*
import com.witelokk.music.tables.*
import com.witelokk.music.tables.Releases
import io.github.smiley4.ktoropenapi.get
import io.ktor.http.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.routing.*
import io.ktor.server.response.*
import org.jetbrains.exposed.sql.Random
import org.jetbrains.exposed.sql.select
import org.jetbrains.exposed.sql.selectAll
import org.jetbrains.exposed.sql.transactions.transaction
import java.time.LocalDate
import java.time.ZoneOffset
import java.util.*

fun Route.homeScreenLayoutRoutes() {
    authenticate("auth-jwt") {
        get("/home-screen-layout", {
            description = "Get home screen layout"
            tags("home-screen")
            response {
                HttpStatusCode.OK to {
                    description = "User's home screen layout"
                    body<HomeScreenLayout>()
                }
            }
        }) {
            val principal = call.principal<JWTPrincipal>()
            val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

            var playlists: List<PlaylistSummary> = listOf()
            var followedArtists: List<ArtistSummary> = listOf()
            var popularReleases: List<Release> = listOf()
            var forYouReleases: List<Release> = listOf()
            var exploreReleases: List<Release> = listOf()

            transaction {
                playlists =
                    Playlists.select { Playlists.userId eq userId }
                        .orderBy(Playlists.createdAt, org.jetbrains.exposed.sql.SortOrder.DESC)
                        .map {
                            PlaylistSummary(
                                id = it[Playlists.id],
                                name = it[Playlists.name],
                                coverUrl = null,
                                songsCount = PlaylistSongs.select { PlaylistSongs.playlistId eq it[Playlists.id] }.count().toInt(),
                            )
                        }
                followedArtists = Followers.leftJoin(Artists).select { Followers.userId eq userId }.map {
                    ArtistSummary(
                        id = it[Artists.id],
                        name = it[Artists.name],
                        avatarUrl = it[Artists.avatarUrl]
                    )
                }
                val todaySeed = (LocalDate.now(ZoneOffset.UTC).toEpochDay() % 1000) / 1000.0 // value between 0 and 1
                exec("SELECT setseed($todaySeed)")
                forYouReleases = Releases.selectAll().orderBy(Random()).limit(4).map {
                    getReleaseWithArtist(userId, it[Releases.id])!!
                }
                exec("SELECT setseed(${(todaySeed + 0.1) % 1.0})")
                popularReleases = Releases.selectAll().orderBy(Random()).limit(4).map {
                    getReleaseWithArtist(userId, it[Releases.id])!!
                }
                exec("SELECT setseed(${(todaySeed + 0.2) % 1.0})")
                exploreReleases = Releases.selectAll().orderBy(Random()).limit(4).map {
                    getReleaseWithArtist(userId, it[Releases.id])!!
                }
            }

            call.respond(
                HttpStatusCode.OK, HomeScreenLayout(
                    playlists = PlaylistsSummary(
                        count = playlists.size,
                        playlists = playlists,
                    ),
                    followedArtists = ArtistsSummary(
                        count = followedArtists.size,
                        artists = followedArtists,
                    ),
                    sections = listOf(
                        HomeScreenLayout.Companion.Section(
                            title = "For you",
                            titleRu = "Для вас",
                            releases = com.witelokk.music.models.Releases(
                                count = forYouReleases.size,
                                releases = forYouReleases,
                            )
                        ),
                        HomeScreenLayout.Companion.Section(
                            title = "Popular",
                            titleRu = "Популярное",
                            releases = com.witelokk.music.models.Releases(
                                count = popularReleases.size,
                                releases = popularReleases,
                            )
                        ),
                        HomeScreenLayout.Companion.Section(
                            title = "Explore",
                            titleRu = "Откройте новое",
                            releases = com.witelokk.music.models.Releases(
                                count = exploreReleases.size,
                                releases = exploreReleases,
                            )
                        ),
                    )
                )
            )
        }
    }
}