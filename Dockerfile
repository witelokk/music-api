FROM openjdk:17-jdk-slim AS build

WORKDIR /app

COPY build.gradle.kts settings.gradle.kts gradlew /app/

COPY gradle /app/gradle

RUN ./gradlew --no-daemon clean build

COPY . /app/

RUN ./gradlew clean build -x test

FROM openjdk:17-jdk-slim

WORKDIR /app

COPY --from=build /app/build/libs/*.jar app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "app.jar"]
