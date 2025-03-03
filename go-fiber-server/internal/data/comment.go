package data

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go-fiber-ent-web-layout/internal/usercase"
	sqlbuild "go-fiber-ent-web-layout/pkg/sql-build"
	"log/slog"
	"strconv"
	"time"
)

type CommentRepo struct {
	db *pgxpool.Pool
}

func NewCommentRepo(data *Data) usercase.ICommentRepo {
	return &CommentRepo{
		db: data.Db,
	}
}

func (self *CommentRepo) Save(comment *usercase.Comment) error {
	builder := sqlbuild.NewInsertBuilder("t_blog_comment").
		Fields("content", "user_id", "fid", "rid", "location", "comment_ip", "comment_ua", "comment_type").
		Values(comment.Content, comment.UserId, comment.Fid, comment.Rid, comment.Location, comment.CommentIp,
			comment.CommentUa, comment.CommentType).
		InsertByCondition(comment.TopicId != nil, "topic_id", comment.TopicId).
		InsertByCondition(comment.ArticleId != nil, "article_id", comment.ArticleId).
		Returning("comment_id")
	var commentId int64
	row := self.db.QueryRow(context.Background(), builder.Sql(), builder.Args()...)
	err := row.Scan(&commentId)
	if err == nil {
		comment.CommentId = commentId
	}
	return err
}

func (self *CommentRepo) Page(query *usercase.CommentQueryForm) (*usercase.PageData[usercase.CommentVo], error) {
	builder := sqlbuild.NewSelectBuilder("t_blog_comment as bc").
		Select("bc.comment_id", "bc.content", "bc.user_id", "bc.fid", "bc.rid", "bc.location", "bc.comment_ua", "bc.vote_up", "bc.comment_type", "bc.is_hot", "bc.is_top", "bc.is_coll", "bc.create_time").
		LeftJoin("t_blog_user as bu").On("bu.user_id").EqRaw("bc.user_id").BuildAsSelect().
		LeftJoin("t_blog_user_extend as bue").On("bue.user_id").EqRaw("bc.user_id").BuildAsSelect().
		Select("jsonb_build_object('nickname', bu.nick_name, 'avatar', bu.avatar, 'level', bue.level, 'labels', bu.labels, 'link', bu.link) as user").
		LeftJoin("t_blog_comment as rbc").On("rbc.comment_id").EqRaw("bc.rid").And("bc.rid").LeRaw("0").BuildAsSelect().
		LeftJoin("t_blog_user as rbu").On("rbu.user_id").EqRaw("rbc.user_id").BuildAsSelect().
		LeftJoin("t_blog_user_extend as rbue").On("rbue.user_id").EqRaw("rbc.user_id").BuildAsSelect().
		Select("CASE WHEN bc.rid > 0 THEN jsonb_build_object('nickname', rbu.nick_name, 'avatarUrl', rbu.avatar, 'level', rbue.level, 'labels', rbu.labels, 'link', rbu.link) END as parentUser").
		WhereByCondition(query.CommentType > 0, "bc.comment_type").Eq(query.CommentType).
		AndByCondition(query.ArticleId > 0, "bc.article_id").Eq(query.ArticleId).
		AndByCondition(query.TopicId > 0, "bc.topic_id").Eq(query.TopicId).
		And("bc.fid").Eq(query.Fid).
		And("bc.delete_at").EqRaw("0").
		And("bc.status").EqRaw("0").BuildAsSelect().
		OrderBy("bc.is_top desc", "bc.sort", "bc.create_time")
	return SelectPage[usercase.CommentVo](builder, query.Page, query.Size, true, self.db)
}

func (self *CommentRepo) TotalComment(query *usercase.CommentQueryForm) (uint64, error) {
	builder := sqlbuild.NewSelectBuilder("t_blog_comment").
		Select("count(*) as total").
		WhereByCondition(query.CommentType > 0, "comment_type").Eq(query.CommentType).
		AndByCondition(query.ArticleId > 0, "article_id").Eq(query.ArticleId).
		AndByCondition(query.TopicId > 0, "topic_id").Eq(query.TopicId).
		And("status").EqRaw("0").
		And("delete_at").EqRaw("0").BuildAsSelect()
	var total uint64
	row := self.db.QueryRow(context.Background(), builder.Sql(), builder.Args()...)
	err := row.Scan(&total)
	return total, err
}

func (self *CommentRepo) ManagePage(query *usercase.CommentQueryForm) (*usercase.PageData[usercase.CommentManageVo], error) {
	builder := sqlbuild.NewSelectBuilder("t_blog_comment as bc").
		Select("bc.*").
		LeftJoin("t_blog_user as bu").On("bu.user_id").EqRaw("bc.user_id").BuildAsSelect().
		Select("bu.username as username").
		LeftJoin("t_blog_article as ba").On("ba.article_id").EqRaw("bc.article_id").BuildAsSelect().
		Select("ba.title as article_title").
		WhereByCondition(query.CommentType > 0, "bc.comment_type").Eq(query.CommentType).
		AndByCondition(query.ArticleId > 0, "bc.article_id").Eq(query.ArticleId).
		AndByCondition(query.TopicId > 0, "bc.topic_id").Eq(query.TopicId).
		AndByCondition(query.Fid > 0, "bc.fid").Eq(query.Fid).
		AndByCondition(query.CreateTimeBegin != "", "bc.create_time").Ge(query.CreateTimeBegin).
		AndByCondition(query.CreateTimeEnd != "", "bc.create_time").Le(query.CreateTimeEnd).
		And("bc.delete_at").EqRaw("0").BuildAsSelect().
		OrderBy("bc.create_time desc", "bc.sort")
	return SelectPage[usercase.CommentManageVo](builder, query.Page, query.Size, true, self.db)
}

func (self *CommentRepo) UpdateSelective(form *usercase.CommentUpdateForm) error {
	builder := sqlbuild.NewUpdateBuilder("t_blog_comment").
		SetRaw("update_time", "now()").
		SetByCondition(form.IsHot != nil, "is_hot", form.IsHot).
		SetByCondition(form.IsTop != nil, "is_top", form.IsTop).
		SetByCondition(form.IsColl != nil, "is_coll", form.IsColl).
		SetByCondition(form.Status != nil, "status", form.Status).
		Where("comment_id").Eq(form.CommentId).BuildAsUpdate()
	result, err := self.db.Exec(context.Background(), builder.Sql(), builder.Args()...)
	if err == nil {
		slog.Info("快捷更新评论完成", "rows", result.RowsAffected(), "commentId", form.CommentId)
	}
	return err
}

func (self *CommentRepo) DeleteById(commentId int64) error {
	builder := sqlbuild.NewUpdateBuilder("t_blog_comment").
		Set("delete_at", time.Now().UnixMilli()).
		Where("comment_id").Eq(commentId).BuildAsUpdate()
	result, err := self.db.Exec(context.Background(), builder.Sql(), builder.Args()...)
	if err == nil {
		slog.Info("删除评论成功", "rows", result.RowsAffected(), "commentId", commentId)
	}
	return err
}

func (self *CommentRepo) CommentVoteUp(commentId int64, num int) error {
	builder := sqlbuild.NewUpdateBuilder("t_blog_comment").
		SetRaw("update_time", "now()").
		SetRaw("vote_up", "vote_up + "+strconv.Itoa(num)).
		Where("comment_id").Eq(commentId).BuildAsUpdate()
	result, err := self.db.Exec(context.Background(), builder.Sql(), builder.Args()...)
	if err == nil {
		slog.Info("更新评论点赞数成功", "rows", result.RowsAffected(), "commentId", commentId)
	}
	return err
}
