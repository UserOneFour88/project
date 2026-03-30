create table if not exists artists (
  id bigserial primary key,
  name text not null,
  genre text not null,
  description text,
  image_path text null
);

create table if not exists tracks (
  id bigserial primary key,
  artist_id bigint not null references artists(id) on delete cascade,
  name text not null,
  duration text not null,
  sort_order int not null default 0
);

create index if not exists idx_tracks_artist_sort on tracks(artist_id, sort_order);

insert into artists (id, name, genre, description, image_path) values
  (1, 'Ghost', 'Рок', 'Шведская рок-группа из Линчёпинга', 'images/ghost.jpg'),
  (2, 'Katagiri', 'Электронная музыка', 'Электронный продюсер', null),
  (3, 'Aether Realm', 'Метал', 'Американская фолк-метал группа', 'images/aether-realm.jpg'),
  (4, 'Breaking Benjamin', 'Рок', 'Американская рок-группа', 'images/breaking-benjamin.jpg'),
  (5, 'Polyphia', 'Прочее', 'Инструментальный прогрессив-метал, США', 'images/polyphia.jpg'),
  (6, 'System of a Down', 'Метал', 'Армяно-американская альтернативная метал-группа', 'images/system-of-a-down.jpg'),
  (7, 'The Living Tombstone', 'Электронная музыка', 'Электронный проект, известен по фан-трекам', 'images/the-living-tombstone.jpg'),
  (8, 'Theory of a Deadman', 'Рок', 'Канадская рок-группа', 'images/theory-of-a-deadman.jpg');

insert into tracks (artist_id, name, duration, sort_order) values
  (1, 'Mary on a cross', '4:56', 1),
  (1, 'Hunter''s Moon', '4:03', 2),
  (1, 'He is', '3:05', 3),
  (2, 'Tachypsychia', '3:36', 1),
  (2, 'Synthetic heartbeat', '3:12', 2),
  (2, 'Neon drift', '4:01', 3),
  (3, 'The sun The moon The star', '17:28', 1),
  (3, 'The Magician', '4:22', 2),
  (3, 'Guardian', '5:01', 3),
  (4, 'Dance With The Devil', '4:45', 1),
  (4, 'The Diary of Jane', '3:20', 2),
  (4, 'Breath', '3:38', 3),
  (5, 'Playing God', '3:23', 1),
  (5, 'G.O.A.T.', '4:09', 2),
  (5, 'Ego Death', '3:45', 3),
  (6, 'Chop Suey!', '3:30', 1),
  (6, 'Toxicity', '3:38', 2),
  (6, 'Aerials', '3:55', 3),
  (7, 'Discord', '4:10', 1),
  (7, 'Five Nights at Freddy''s', '4:02', 2),
  (7, 'My Ordinary Life', '3:48', 3),
  (8, 'Bad Girlfriend', '3:25', 1),
  (8, 'Rx (Medicate)', '3:02', 2),
  (8, 'Not Meant to Apologize', '3:44', 3);

select setval(pg_get_serial_sequence('artists', 'id'), (select coalesce(max(id), 1) from artists));
select setval(pg_get_serial_sequence('tracks', 'id'), (select coalesce(max(id), 1) from tracks));
