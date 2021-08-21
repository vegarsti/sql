CREATE TABLE genres (id INTEGER, name STRING);
INSERT INTO genres VALUES (1, 'Science Fiction');
INSERT INTO genres VALUES (2, 'Action');
INSERT INTO genres VALUES (3, 'Drama');
INSERT INTO genres VALUES (4, 'Comedy');

CREATE TABLE movies (id INTEGER, title STRING, studio_id INTEGER, genre_id INTEGER, released INTEGER, rating FLOAT);
INSERT INTO movies VALUES (1,  'Stalker', 1, 1, 1979, 8.2);
INSERT INTO movies VALUES (2,  'Sicario', 2, 2, 2015, 7.6);
INSERT INTO movies VALUES (3,  'Primer', 3, 1, 2004, 6.9);
INSERT INTO movies VALUES (4,  'Heat', 4, 2, 1995, 8.2);
INSERT INTO movies VALUES (5,  'The Fountain', 4, 1, 2006, 7.2);
INSERT INTO movies VALUES (6,  'Solaris', 1, 1, 1972, 8.1);
INSERT INTO movies VALUES (7,  'Gravity', 4, 1, 2013, 7.7);
INSERT INTO movies VALUES (8,  '21 Grams', 5, 3, 2003, 7.7);
INSERT INTO movies VALUES (9,  'Birdman', 4, 4, 2014, 7.7);
INSERT INTO movies VALUES (10, 'Inception', 4, 1, 2010, 8.8);
INSERT INTO movies VALUES (11, 'Lost in Translation', 5, 4, 2003, 7.7);
INSERT INTO movies VALUES (12, 'Eternal Sunshine of the Spotless Mind', 5, 3, 2004, 8.3);

SELECT title, rating FROM movies WHERE released >= 2000 ORDER BY rating DESC LIMIT 3;
SELECT movies.id, movies.title, genres.name FROM movies JOIN genres ON movies.genre_id = genres.id LIMIT 4;



CREATE TABLE studios (id INTEGER, name STRING);
INSERT INTO studios VALUES (1, 'Mosfilm');
INSERT INTO studios VALUES (2, 'Lionsgate');
INSERT INTO studios VALUES (3, 'StudioCanal');
INSERT INTO studios VALUES (4, 'Warner Bros');
INSERT INTO studios VALUES (5, 'Focus Features');
