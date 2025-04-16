create table "Crimes" (
	"id_crime" bigint not null,
	"nome_crime" varchar(255) null,
	"peso_crime" smallint null
);

alter table
	"Crimes" add primary key("id_crime");


create table "Denuncia"(
	"id_denuncia" bigint not null,
	"id_crime" bigint not null,
	"id_bairro" bigint not null,
	"data" timestamp(0) without time zone not null
);
alter table 
	"Denuncia" add primary key("id_denuncia");

create table "Bairros"(
	"id_bairro" bigint not null,
	"latitude" varchar(255) not null,
	"longitude" varchar(255) not null,
	"nome" varchar(255) not null,
	"peso_bairro" bigint not null
);
alter table 
	"Bairros" add primary key("id_bairro");

alter table 
	"Denuncia" add constraint "denuncia_id_crime_foreign" foreign key("id_crime") references "Crimes"("id_crime");

alter table 
	"Denuncia" add constraint "denuncia_id_bairro_foreign" foreign key("id_bairro") references "Bairros"("id");


