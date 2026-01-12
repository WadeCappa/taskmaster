create database taskmaster_db;

\c taskmaster_db

create table if not exists tasks (
	task_id bigint,
	user_id bigint,
	fields jsonb,
	priority smallint,
	status smallint,
    
	primary key (task_id)
);
-- we'll be doing queries on this, where priority may not be specified
CREATE index if not exists task_lookup on tasks (user_id, status, priority);
create sequence if not exists task_ids start 101;

create table if not exists tags (
	user_id bigint,
	tag_id bigint,
	write_time timestamptz,
	name varchar(256),

	primary key (tag_id)
);
-- we need to lookup tags by tag id, so we need this index
create index if not exists tag_lookup on tags (name);
create sequence if not exists tag_ids start 101;

create table if not exists tags_to_tasks (
	task_id bigint,
	tag_id bigint,

	primary key (tag_id, task_id),
	foreign key (task_id) references tasks(task_id) on delete cascade,
	foreign key (tag_id) references tags(tag_id) on delete cascade
);
-- need this to figure out the tags on a task when we return it
create index if not exists tasks_to_tags_index on tags_to_tasks (task_id, tag_id);

create table if not exists addendums (
	addendum_id bigint,
	user_id bigint,
	task_id bigint,
	content text,
	write_time timestamptz,

	primary key (addendum_id),
	foreign key (task_id) references tasks(task_id) on delete cascade
);
-- sorting by write_time so that we can return addendums in the correct order
create index if not exists addendum_lookup on addendums (task_id, write_time);
create sequence if not exists addendum_ids start 101;
