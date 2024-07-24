-- Deletar registros das tabelas
DELETE
FROM public.user_group_permission;
DELETE
FROM public.permission;
DELETE
FROM public.user_group;
DELETE
FROM public.user;

-- Remover tabelas
DROP TABLE public.user_group_permission;
DROP TABLE public.user;
DROP TABLE public.user_group;
DROP TABLE public.permission;