create table "Crimes" (
	"crime_id" bigint not null,
	"crime_name" varchar(255) null,
	"crime_weight" smallint null
);

alter table
	"Crimes" add primary key("crime_id");


create table "Reports"(
	"report_id" bigint not null,
	"crime_id" bigint not null,
	"neighborhood_id" bigint not null,
	"report_date" timestamp(0) without time zone not null
);
alter table 
	"Reports" add primary key("report_id");

create table "Neighborhoods"(
	"neighborhood_id" bigint not null,
	"latitude" varchar(255) not null,
	"longitude" varchar(255) not null,
	"name" varchar(255) not null,
	"neighborhood_weight" bigint not null
);
alter table 
	"Neighborhoods" add primary key("neighborhood_id");

alter table 
	"Reports" add constraint "crime_foreign_id" foreign key("crime_id") references "Crimes"("crime_id");

alter table 
	"Reports" add constraint "neighborhood_foreign_id" foreign key("neighborhood_id") references "Neighborhoods"("neighborhood_id");


