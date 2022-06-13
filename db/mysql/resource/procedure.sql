DELIMITER &&
CREATE DEFINER=`root`@`localhost` PROCEDURE `list_objects`(keyid bigint(20) ,
  pfix varchar(1000) ,
  dlmt varchar(10) ,
  lmt int(10) ,
  marker varchar(1000) )
begin 
SET SESSION SQL_LOG_BIN = 0;
drop temporary table if exists tmp_path ;
create temporary table tmp_path (path varchar(1000))  engine = memory ;
lets : begin
declare pfix_r varchar(1000) ;
declare cnt_rs int(10) default 0 ;
declare cnt_rpt_3 int(10) default 0 ;
declare v_pfx_lth int(10) ;
declare v_dlt_lth int(10) ;
declare v_path varchar(1000) ;
declare v_hv_dlmt int(10) ;
declare v_tmp varchar(1000) default null ;
declare v_tmp_pure varchar(1000) default null ;
declare v_path_next varchar(1000) ;
declare done int default 0 ;
declare cur_1_nomark cursor for
    select substring_index(substring(object_name,v_pfx_lth+1),dlmt,1),locate(dlmt,substring(object_name,char_length(concat(pfix,substring_index(substring(object_name,v_pfx_lth+1),dlmt,1)))+1))
    from object_lists
    where bucket_id = keyid and object_name like concat(pfix_r,'%') limit 100 ;
declare cur_1_mark_file cursor for
    select substring_index(substring(object_name,v_pfx_lth+1),dlmt,1),locate(dlmt,substring(object_name,char_length(concat(pfix,substring_index(substring(object_name,v_pfx_lth+1),dlmt,1)))+1))
    from object_lists
    where bucket_id = keyid and object_name like concat(pfix_r,'%') and object_name > marker limit 100 ;
declare cur_1_mark_path cursor for
    select substring_index(substring(object_name,v_pfx_lth+1),dlmt,1),locate(dlmt,substring(object_name,char_length(concat(pfix,substring_index(substring(object_name,v_pfx_lth+1),dlmt,1)))+1))
    from object_lists
    where bucket_id = keyid and object_name like concat(pfix_r,'%') and object_name > concat(marker,x'EFBFBF') limit 100 ;
declare cur_2_skip cursor for
    select substring_index(substring(object_name,v_pfx_lth+1),dlmt,1),locate(dlmt,substring(object_name,char_length(concat(pfix,substring_index(substring(object_name,v_pfx_lth+1),dlmt,1)))+1))
    from object_lists
    where bucket_id = keyid and object_name like concat(pfix_r,'%') and object_name > concat(concat(pfix,v_tmp_pure),x'EFBFBF') limit 100 ;
declare cur_2_coskip cursor for
    select substring_index(substring(object_name,v_pfx_lth+1),dlmt,1),locate(dlmt,substring(object_name,char_length(concat(pfix,substring_index(substring(object_name,v_pfx_lth+1),dlmt,1)))+1))
    from object_lists
    where bucket_id = keyid and object_name like concat(pfix_r,'%') and object_name > concat(concat(pfix,v_tmp_pure)) limit 100 ;
declare cur_nodlmt_nomark cursor for
    select object_name
    from object_lists
    where bucket_id = keyid and object_name like concat(pfix_r,'%') limit lmt ;
declare cur_nodlmt_mark cursor for
    select object_name
    from object_lists
    where bucket_id = keyid and object_name like concat(pfix_r,'%') and object_name > marker limit lmt ;
declare continue handler for sqlstate '02000' set done = 1;
select replace(replace(pfix,'_','\_'),'%','\%') into pfix_r ;
select char_length(pfix) into v_pfx_lth ;
select char_length(dlmt) into v_dlt_lth ;
set v_tmp = null ;
if v_dlt_lth = 0 then
    if marker is null then 
        open cur_nodlmt_nomark ;
        rpt_0 : repeat
            fetch cur_nodlmt_nomark into v_path ;
            if done = 1 then
                leave rpt_0 ;
            end if ;
            insert into tmp_path (path) values (v_path) ;
        until done = 1 end repeat rpt_0 ;
        close cur_nodlmt_nomark ;
    else 
        open cur_nodlmt_mark ;
        rpt_0 : repeat
            fetch cur_nodlmt_mark into v_path ;
            if done = 1 then
                leave rpt_0 ;
            end if ;
            insert into tmp_path (path) values (v_path) ;
        until done = 1 end repeat ;
        close cur_nodlmt_mark ;
    end if ;
    leave lets ;
end if ;
if marker is null then 
    open cur_1_nomark ;
    rpt_1_nomark : repeat
        fetch cur_1_nomark into v_path,v_hv_dlmt ;
        if done = 1 then
            leave rpt_1_nomark ;
        end if ;
        if v_tmp is null then
            if  v_hv_dlmt = 1 then
                insert into tmp_path (path) values (concat(pfix,v_path,dlmt)) ;
                set v_tmp = concat(pfix,v_path,dlmt) ;
            else 
                insert into tmp_path (path) values (concat(pfix,v_path)) ;
                set v_tmp = concat(pfix,v_path) ;
            end if ;
            set cnt_rs = cnt_rs +1 ;
        elseif v_tmp is not null and concat(pfix,v_path,dlmt) <> v_tmp  and v_hv_dlmt = 1 then
            insert into tmp_path (path) values (concat(pfix,v_path,dlmt)) ;
            set v_tmp = concat(pfix,v_path,dlmt) ;
            set cnt_rs = cnt_rs +1 ;
        elseif v_tmp is not null and concat(pfix,v_path) <> v_tmp  and v_hv_dlmt <> 1 then
            insert into tmp_path (path) values (concat(pfix,v_path)) ;
            set v_tmp = concat(pfix,v_path) ;
            set cnt_rs = cnt_rs +1 ;
        elseif v_tmp is not null and concat(pfix,v_path,dlmt) = v_tmp and v_hv_dlmt = 1 then
            set v_tmp = concat(pfix,v_path,dlmt) ;
        elseif v_tmp is not null and concat(pfix,v_path) = v_tmp and v_hv_dlmt <> 1 then
            set v_tmp = concat(pfix,v_path) ;
        end if ;
        if cnt_rs >= lmt then
            close cur_1_nomark ;
            leave lets;
        end if ;
    until done = 1 end repeat rpt_1_nomark ;
    close cur_1_nomark ;
else
    if right(marker,v_dlt_lth) <> dlmt then
        open cur_1_mark_file ;
        rpt_1_mark : repeat
            fetch cur_1_mark_file into v_path,v_hv_dlmt ;
            if done = 1 then
                leave rpt_1_mark ;
            end if ;
            if v_tmp is null then
                if  v_hv_dlmt = 1 then
                    insert into tmp_path (path) values (concat(pfix,v_path,dlmt)) ;
                    set v_tmp = concat(pfix,v_path,dlmt) ;
                else 
                    insert into tmp_path (path) values (concat(pfix,v_path)) ;
                    set v_tmp = concat(pfix,v_path) ;
                end if ;
                set cnt_rs = cnt_rs +1 ;
            elseif v_tmp is not null and concat(pfix,v_path,dlmt) <> v_tmp  and v_hv_dlmt = 1 then
                insert into tmp_path (path) values (concat(pfix,v_path,dlmt)) ;
                set v_tmp = concat(pfix,v_path,dlmt) ;
                set cnt_rs = cnt_rs +1 ;
            elseif v_tmp is not null and concat(pfix,v_path) <> v_tmp  and v_hv_dlmt <> 1 then
                insert into tmp_path (path) values (concat(pfix,v_path)) ;
                set v_tmp = concat(pfix,v_path) ;
                set cnt_rs = cnt_rs +1 ;
            elseif v_tmp is not null and concat(pfix,v_path,dlmt) = v_tmp and v_hv_dlmt = 1 then
                set v_tmp = concat(pfix,v_path,dlmt) ;
            elseif v_tmp is not null and concat(pfix,v_path) = v_tmp and v_hv_dlmt <> 1 then
                set v_tmp = concat(pfix,v_path) ;
            end if ;
            if cnt_rs >= lmt then
                close cur_1_mark_file ;
                leave lets ;
            end if ;
        until done = 1 end repeat rpt_1_mark ;
        close cur_1_mark_file ;
    else 
        open cur_1_mark_path ;
        rpt_1_mark : repeat
            fetch cur_1_mark_path into v_path,v_hv_dlmt ;
            if done = 1 then
                leave rpt_1_mark ;
            end if ;
            if v_tmp is null then
                if  v_hv_dlmt = 1 then
                    insert into tmp_path (path) values (concat(pfix,v_path,dlmt)) ;
                    set v_tmp = concat(pfix,v_path,dlmt) ;
                else 
                    insert into tmp_path (path) values (concat(pfix,v_path)) ;
                    set v_tmp = concat(pfix,v_path) ;
                end if ;
                set cnt_rs = cnt_rs +1 ;
            elseif v_tmp is not null and concat(pfix,v_path,dlmt) <> v_tmp  and v_hv_dlmt = 1 then
                insert into tmp_path (path) values (concat(pfix,v_path,dlmt)) ;
                set v_tmp = concat(pfix,v_path,dlmt) ;
                set cnt_rs = cnt_rs +1 ;
            elseif v_tmp is not null and concat(pfix,v_path) <> v_tmp  and v_hv_dlmt <> 1 then
                insert into tmp_path (path) values (concat(pfix,v_path)) ;
                set v_tmp = concat(pfix,v_path) ;
                set cnt_rs = cnt_rs +1 ;
            elseif v_tmp is not null and concat(pfix,v_path,dlmt) = v_tmp and v_hv_dlmt = 1 then
                set v_tmp = concat(pfix,v_path,dlmt) ;
            elseif v_tmp is not null and concat(pfix,v_path) = v_tmp and v_hv_dlmt <> 1 then
                set v_tmp = concat(pfix,v_path) ;
            end if ;
            if cnt_rs >= lmt then
                close cur_1_mark_path ;
                leave lets ;
            end if ;
        until done = 1 end repeat rpt_1_mark ;
        close cur_1_mark_path ;
    end if ;
end if ; 
if v_tmp is null then
    leave lets ;
end if ;
set done = 0 ;
rpt_2 : repeat
    if  v_hv_dlmt = 1 then
        set v_tmp_pure = substring(v_tmp,v_pfx_lth+1) ;
        open cur_2_skip ;
        rpt_3 : repeat
            fetch cur_2_skip INTO v_path_next,v_hv_dlmt;
            if done = 1 then
                close cur_2_skip ;
                leave rpt_3 ;
            end if ;
            if v_tmp is not null and concat(pfix,v_path_next,dlmt) <> v_tmp  and v_hv_dlmt = 1 then
                insert into tmp_path (path) values (concat(pfix,v_path_next,dlmt)) ;
                set v_tmp = concat(pfix,v_path_next,dlmt) ;
                set cnt_rs = cnt_rs +1 ;
                set cnt_rpt_3 = cnt_rpt_3 +1 ;
            elseif v_tmp is not null and concat(pfix,v_path_next) <> v_tmp  and v_hv_dlmt <> 1 then
                insert into tmp_path (path) values (concat(pfix,v_path_next)) ;
                set v_tmp = concat(pfix,v_path_next) ;
                set cnt_rs = cnt_rs +1 ;
                set cnt_rpt_3 = cnt_rpt_3 +1 ;
            elseif v_tmp is not null and concat(pfix,v_path_next,dlmt) = v_tmp and v_hv_dlmt = 1 then
                set v_tmp = concat(pfix,v_path_next,dlmt) ;
                set cnt_rpt_3 = cnt_rpt_3 +1 ;
            elseif v_tmp is not null and concat(pfix,v_path_next) = v_tmp and v_hv_dlmt <> 1 then
                set v_tmp = concat(pfix,v_path_next) ;
                set cnt_rpt_3 = cnt_rpt_3 +1 ;
            end if ;
            if cnt_rs >= lmt then
                close cur_2_skip ;
                leave lets ;
            end if ;
        until done = 1 end repeat rpt_3;
        if  v_hv_dlmt = 1 then
            set v_tmp_pure = substring(v_tmp,v_pfx_lth+1) ;
        else 
            set v_tmp_pure = substring(v_tmp,v_pfx_lth+1) ;
        end if ;
        if cnt_rpt_3 = 0 then
            leave lets ;
        end if;
        set cnt_rpt_3 = 0 ;
        set done = 0 ;
    else 
        set v_tmp_pure = substring(v_tmp,v_pfx_lth+1) ;
        open cur_2_coskip ;
        rpt_3 : repeat
            fetch cur_2_coskip INTO v_path_next,v_hv_dlmt;
            if done = 1 then
                close cur_2_coskip ;
                leave rpt_3 ;
            end if ;
            if v_tmp is not null and concat(pfix,v_path_next,dlmt) <> v_tmp  and v_hv_dlmt = 1 then
                insert into tmp_path (path) values (concat(pfix,v_path_next,dlmt)) ;
                set v_tmp = concat(pfix,v_path_next,dlmt) ;
                set cnt_rs = cnt_rs +1 ;
                set cnt_rpt_3 = cnt_rpt_3 +1 ;
            elseif v_tmp is not null and concat(pfix,v_path_next) <> v_tmp  and v_hv_dlmt <> 1 then
                insert into tmp_path (path) values (concat(pfix,v_path_next)) ;
                set v_tmp = concat(pfix,v_path_next) ;
                set cnt_rs = cnt_rs +1 ;
                set cnt_rpt_3 = cnt_rpt_3 +1 ;
            elseif v_tmp is not null and concat(pfix,v_path_next,dlmt) = v_tmp and v_hv_dlmt = 1 then
                set v_tmp = concat(pfix,v_path_next,dlmt) ;
                set cnt_rpt_3 = cnt_rpt_3 +1 ;
            elseif v_tmp is not null and concat(pfix,v_path_next) = v_tmp and v_hv_dlmt <> 1 then
                set v_tmp = concat(pfix,v_path_next) ;
                set cnt_rpt_3 = cnt_rpt_3 +1 ;
            end if ;
            if cnt_rs >= lmt then
                close cur_2_coskip ;
                leave lets ;
            end if ;
        until done = 1 end repeat rpt_3;
        if  v_hv_dlmt = 1 then
            set v_tmp_pure = substring(v_tmp,v_pfx_lth+1) ;
        else 
            set v_tmp_pure = substring(v_tmp,v_pfx_lth+1) ;
        end if ;
        if cnt_rpt_3 = 0 then
            leave lets ;
        end if;
        set cnt_rpt_3 = 0 ;
        set done = 0 ;
    end if ;
until done = 65535 end repeat rpt_2 ;
end lets ; 
select path from tmp_path ;
drop temporary table if exists tmp_path ;
SET SESSION SQL_LOG_BIN = 1;
end &&
