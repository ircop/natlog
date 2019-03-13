create database natlog;

create table natlog.connections ( \
 date Date, \
 time DateTime, \
 dst_ip UInt32, \
 nat_ip UInt32, \
 local_ip UInt32, \
 dst_port UInt16, \
 nat_port UInt16, \
 local_port UInt16, \
 proto Enum8('UNKNOWN' = 0, 'TCP' = 1, 'UDP' = 2, 'ICMP' = 3), \
 type FixedString(2) \
) engine=MergeTree(date, (dst_ip, nat_ip, local_ip), 8192)

