SELECT t.*
FROM tasks t
join tags_to_tasks ttt on t.task_id = ttt.task_id
join tags tg on tg.tag_id = ttt.tag_id
where t.user_id = 101
and tg.name = any (array['test-tag','another-test'])
group by t.tasK_id 
having count(distinct tg.tag_id) = cardinality(array['test-tag','another-test']) 
order by priority;

select distinct t.* from tasks t 
JOIN tags_to_tasks ttt ON t.task_id = ttt.task_id 
JOIN tags tg ON tg.tag_id = ttt.tag_id 
where t.user_id = 101 
AND t.priority = 101 
AND tg.name = ANY (array['some-tag', 'some-other-tag']);

insert into tags (user_id, tag_id, write_time, name) values (101, nextval('tag_ids'), NOW(), 'another-test');

select t.*
from tasks t
where t.task_id = $1
and t.user_id = $2

select 
