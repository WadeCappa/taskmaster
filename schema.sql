create database taskmaster_db;

\c taskmaster_db
create sequence task_ids start 101;
create table tasks (
	task_id           bigint primary key default nextval('task_ids'),
	user_id           bigint not null,
	parent_task_id    bigint references tasks(task_id) on delete cascade,
	title             text not null,
	description       text not null default '',
	status            smallint not null default 0,  -- 0=open, 1=in_progress, 2=closed
	created_time      timestamptz not null default now()
);
create index status_task_lookup on tasks (user_id, status);
create index task_by_parent on tasks (parent_task_id);

create table task_blocked_by (
	task_id     bigint not null references tasks(task_id) on delete cascade,
	blocked_by  bigint not null references tasks(task_id) on delete cascade,
	primary key (task_id, blocked_by)
);

create sequence addendum_ids start 101;
create table addendums (
	addendum_id  bigint primary key default nextval('addendum_ids'),
	user_id      bigint not null,
	task_id      bigint not null references tasks(task_id) on delete cascade,
	content      text not null,
	created_time timestamptz not null default now()
);
create index addendum_by_task on addendums (task_id, created_time);

-- The nil task
insert into tasks (task_id, user_Id, parent_task_id, title, description, status, created_time) values (0,0,0,'','',0,now());
