package data

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go-fiber-ent-web-layout/internal/tools"
	"go-fiber-ent-web-layout/internal/usercase"
	sqlbuild "go-fiber-ent-web-layout/pkg/sql-build"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

type ArticleRepo struct {
	db *pgxpool.Pool
}

func NewArticleRepo(data *Data) usercase.IArticleRepo {
	return &ArticleRepo{
		db: data.Db,
	}
}

func (self *ArticleRepo) Save(article *usercase.Article) error {
	builder := sqlbuild.NewInsertBuilder("t_blog_article").
		Fields("title", "summary", "cover_url", "category_ids", "tag_ids", "content", "word_count", "protocol",
			"tips", "password", "is_hot", "is_top", "is_comment", "is_private", "sort", "status").
		Values(article.Title, article.Summary, article.CoverUrl, article.CategoryIds, article.TagIds, article.Content,
			article.WordCount, article.Protocol, article.Tips, article.Password, article.IsHot, article.IsTop,
			article.IsComment, article.IsPrivate, *article.Sort, *article.Status).
		Returning("article_id")
	row := self.db.QueryRow(context.Background(), builder.Sql(), builder.Args()...)
	var articleId uint64
	err := row.Scan(&articleId)
	if err == nil {
		slog.Info("保存博客文章完成", "articleId", articleId)
		article.ArticleId = articleId
	}
	return err
}

func (self *ArticleRepo) Update(article *usercase.Article) error {
	builder := sqlbuild.NewUpdateBuilder("t_blog_article").
		SetRaw("update_time", "now()").
		SetByMap(map[string]any{
			"title":        article.Title,
			"summary":      article.Summary,
			"cover_url":    article.CoverUrl,
			"category_ids": article.CategoryIds,
			"tag_ids":      article.TagIds,
			"protocol":     article.Protocol,
			"tips":         article.Tips,
			"password":     article.Password,
			"is_hot":       article.IsHot,
			"is_top":       article.IsTop,
			"is_comment":   article.IsComment,
			"is_private":   article.IsPrivate,
			"sort":         article.Sort,
			"status":       article.Status,
		})
	if strings.TrimSpace(article.Content) != "" {
		builder.Set("content", article.Content)
		builder.Set("word_count", article.WordCount)
	}
	builder.Where("article_id").Eq(article.ArticleId).BuildAsUpdate()
	result, err := self.db.Exec(context.Background(), builder.Sql(), builder.Args()...)
	if err == nil {
		slog.Info("更新博客文章完成", "row", result.RowsAffected(), "articleId", article.ArticleId)
	}
	return err
}

func (self *ArticleRepo) UpdateSelective(form *usercase.ArticleUpdateForm) error {
	builder := sqlbuild.NewUpdateBuilder("t_blog_article").
		SetRaw("update_time", "now()").
		SetByCondition(form.IsHot != nil, "is_hot", form.IsHot).
		SetByCondition(form.IsTop != nil, "is_top", form.IsTop).
		SetByCondition(form.IsComment != nil, "is_comment", form.IsComment).
		SetByCondition(form.Status != nil, "status", form.Status).
		Where("article_id").Eq(form.ArticleId).BuildAsUpdate()
	result, err := self.db.Exec(context.Background(), builder.Sql(), builder.Args()...)
	if err == nil {
		slog.Info("快捷更新博客文章完成", "row", result.RowsAffected(), "articleId", form.ArticleId)
	}
	return err
}

func (self *ArticleRepo) UpdateViewNum(articleId uint64, num int) error {
	builder := sqlbuild.NewUpdateBuilder("t_blog_article").
		SetRaw("view_num", "view_num + "+strconv.Itoa(num)).
		Where("article_id").Eq(articleId).BuildAsUpdate()
	result, err := self.db.Exec(context.Background(), builder.Sql(), articleId)
	if err == nil {
		slog.Info("更新博客文章查看次数完成", "rows", result.RowsAffected(), "articleId", articleId)
	}
	return err
}

func (self *ArticleRepo) Page(query *usercase.ArticleQueryForm) ([]*usercase.ArticleVo, int64, error) {
	var articleSelectFields []string
	if query.IsAdmin {
		articleSelectFields = []string{"ba.article_id", "ba.title", "ba.summary", "ba.cover_url", "ba.category_ids",
			"ba.tag_ids", "ba.view_num", "ba.share_num", "ba.vote_up", "ba.protocol", "ba.tips", "ba.password", "ba.is_hot",
			"ba.is_top", "ba.is_comment", "ba.is_private", "ba.create_time", "ba.sort", "ba.status", "ba.word_count"}
	} else {
		articleSelectFields = []string{"ba.article_id", "ba.title", "ba.summary", "ba.cover_url", "ba.view_num",
			"ba.share_num", "ba.vote_up", "ba.is_hot", "ba.is_top", "ba.create_time", "ba.word_count"}
	}
	builder := sqlbuild.NewSelectBuilder("t_blog_article as ba").
		Select(articleSelectFields...).
		CountField("ba.article_id").
		LeftJoin("t_blog_comment as bc").On("bc.article_id").EqRaw("ba.article_id").And("bc.status").EqRaw("0").And("bc.delete_at").EqRaw("0").BuildAsSelect().
		Select("count(DISTINCT bc.comment_id) as comment_num").
		LeftJoin("t_blog_category as ct").On("ct.category_id").EqRaw("ANY(ba.category_ids)").And("ct.status").EqRaw("0").And("ct.delete_at").EqRaw("0").BuildAsSelect().
		// 直接使用 jsonb_agg 将分类字段聚合为分类列表
		// json字段需要使用驼峰命名 因为pgx在处理jsonb类型字段时 会调用json库直接反序列化
		Select("jsonb_agg(DISTINCT jsonb_build_object('categoryId', ct.category_id, 'categoryName', ct.category_name)) AS categories").
		LeftJoin("t_blog_tag as bt").On("bt.tag_id").EqRaw("ANY(ba.tag_ids)").And("bt.status").EqRaw("0").And("bt.delete_at").EqRaw("0").BuildAsSelect().
		// 聚合为标签列表
		Select("jsonb_agg(DISTINCT jsonb_build_object('tagId', bt.tag_id, 'tagName', bt.tag_name, 'color', bt.color)) AS tags").
		WhereByCondition(query.Title != "", "ba.title").Like("%"+query.Title+"%").
		AndByCondition(query.TagId > 0, fmt.Sprintf("%d", query.TagId)).EqRaw("ANY(ba.tag_ids)").
		AndByCondition(query.CategoryId > 0, fmt.Sprintf("%d", query.CategoryId)).EqRaw("ANY(ba.category_ids)").
		AndByCondition(!query.IsAdmin, "ba.is_top").EqRaw("false").
		AndByCondition(query.CreateTimeBegin != "", "ba.create_time").Ge(query.CreateTimeBegin).
		AndByCondition(query.CreateTimeEnd != "", "ba.create_time").Le(query.CreateTimeEnd).
		AndByCondition(query.Status != nil, "ba.status").Eq(query.Status).
		And("ba.delete_at").EqRaw("0").BuildAsSelect().
		GroupBy("ba.article_id").OrderBy("ba.is_top desc", "ba.sort", "ba.create_time desc")
	row := self.db.QueryRow(context.Background(), builder.CountSql(), builder.Args()...)
	var total int64
	if err := row.Scan(&total); err != nil {
		return nil, 0, err
	}
	articles := make([]*usercase.ArticleVo, 0)
	if total == 0 {
		return articles, 0, nil
	}
	offset := tools.ComputeOffset(total, query.Page, query.Size, true)
	builder.Limit(int64(query.Size)).Offset(offset)
	rows, err := self.db.Query(context.Background(), builder.Sql(), builder.Args()...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	articles, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (*usercase.ArticleVo, error) {
		return pgx.RowToAddrOfStructByNameLax[usercase.ArticleVo](row)
	})
	rows.Close()
	return articles, total, err
}

func (self *ArticleRepo) ListTopArticle() ([]*usercase.Article, error) {
	builder := sqlbuild.NewSelectBuilder("t_blog_article").
		Select("article_id", "title", "summary", "cover_url", "view_num", "share_num", "vote_up", "is_hot", "is_top", "create_time", "word_count").
		Where("is_top").EqRaw("true").
		And("status").EqRaw("0").
		And("delete_at").EqRaw("0").BuildAsSelect().
		GroupBy("article_id").
		OrderBy("sort", "create_time desc")
	rows, err := self.db.Query(context.Background(), builder.Sql(), builder.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (*usercase.Article, error) {
		return pgx.RowToAddrOfStructByNameLax[usercase.Article](row)
	})
}

func (self *ArticleRepo) ListHotArticle() ([]usercase.SimpleArticleVo, error) {
	builder := sqlbuild.NewSelectBuilder("t_blog_article").
		Select("article_id", "title").
		Where("status").EqRaw("0").And("delete_at").EqRaw("0").BuildAsSelect().
		OrderByDesc("is_hot", "vote_up", "sort").
		Limit(8)
	rows, err := self.db.Query(context.Background(), builder.Sql())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (usercase.SimpleArticleVo, error) {
		vo := usercase.SimpleArticleVo{}
		scanErr := row.Scan(&vo.ArticleId, &vo.Title)
		return vo, scanErr
	})
}

func (self *ArticleRepo) ListRelatedArticle(articleId uint64, limit int) ([]usercase.SimpleArticleVo, error) {
	builder := sqlbuild.NewSelectBuilder("t_blog_article as ba1").
		Select("ba2.article_id", "ba2.title", "ba2.cover_url").
		InnerJoin("t_blog_article as ba2").On("ba2.article_id").NeRaw("ba1.article_id").
		And("ba2.tag_ids").OriginalRaw("&&", "ba1.tag_ids").
		And("ba2.status").EqRaw("0").
		And("ba2.delete_at").EqRaw("0").BuildAsSelect().
		Where("ba1.article_id").Eq(articleId).
		And("ba1.status").EqRaw("0").
		And("ba1.delete_at").EqRaw("0").BuildAsSelect().
		OrderBy("ba2.sort", "ba2.view_num desc", "ba2.create_time desc").
		Limit(int64(limit))
	rows, err := self.db.Query(context.Background(), builder.Sql(), articleId)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (usercase.SimpleArticleVo, error) {
		vo := usercase.SimpleArticleVo{}
		scanErr := row.Scan(&vo.ArticleId, &vo.Title, &vo.CoverUrl)
		return vo, scanErr
	})
}

func (self *ArticleRepo) PageByLabel(query *usercase.ArticleQueryForm) ([]*usercase.Article, int64, error) {
	builder := sqlbuild.NewSelectBuilder("t_blog_article").
		Select("article_id", "title", "summary", "cover_url", "view_num", "share_num", "vote_up", "is_hot", "is_top", "create_time").
		Where("status").EqRaw("0").
		AndByCondition(query.TagId > 0, fmt.Sprintf("%d", query.TagId)).EqRaw("ANY(tag_ids)").
		AndByCondition(query.CategoryId > 0, fmt.Sprintf("%d", query.CategoryId)).EqRaw("ANY(category_ids)").BuildAsSelect().
		OrderBy("is_top desc", "sort", "create_time desc")
	row := self.db.QueryRow(context.Background(), builder.CountSql(), builder.Args()...)
	var total int64
	if err := row.Scan(&total); err != nil {
		return nil, 0, err
	}
	articles := make([]*usercase.Article, 0)
	if total == 0 {
		return articles, 0, nil
	}
	offset := tools.ComputeOffset(total, query.Page, query.Size, true)
	builder.Limit(int64(query.Size)).Offset(offset)
	rows, err := self.db.Query(context.Background(), builder.Sql(), builder.Args()...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	articles, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (*usercase.Article, error) {
		return pgx.RowToAddrOfStructByNameLax[usercase.Article](row)
	})
	return articles, total, err
}

func (self *ArticleRepo) Archives() ([]usercase.ArticleArchive, error) {
	builder := sqlbuild.NewSelectBuilder("t_blog_article").
		Select("to_char(create_time, 'YYYY-MM') as month", "count(*) as total").
		Where("status").EqRaw("0").And("delete_at").EqRaw("0").BuildAsSelect().
		GroupBy("month").OrderByDesc("month")
	rows, err := self.db.Query(context.Background(), builder.Sql())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (usercase.ArticleArchive, error) {
		return pgx.RowToStructByName[usercase.ArticleArchive](row)
	})
}

func (self *ArticleRepo) SelectById(articleId uint64, isAdmin bool) (*usercase.ArticleVo, error) {
	var selectFields []string
	if isAdmin {
		selectFields = []string{"ba.*"}
	} else {
		selectFields = []string{"ba.article_id", "ba.title", "ba.summary", "ba.cover_url", "ba.view_num", "ba.share_num",
			"ba.content", "ba.protocol", "ba.tips", "ba.is_hot", "ba.is_top", "ba.is_comment", "ba.create_time",
			"ba.update_time", "ba.vote_up", "ba.word_count"}
	}
	builder := sqlbuild.NewSelectBuilder("t_blog_article as ba").
		Select(selectFields...).
		LeftJoin("t_blog_category as ct").On("ct.category_id").EqRaw("ANY(ba.category_ids)").And("ct.status").EqRaw("0").And("ct.delete_at").EqRaw("0").BuildAsSelect().
		Select("jsonb_agg(DISTINCT jsonb_build_object('categoryId', ct.category_id, 'categoryName', ct.category_name)) AS categories").
		LeftJoin("t_blog_tag as bt").On("bt.tag_id").EqRaw("ANY(ba.tag_ids)").And("bt.status").EqRaw("0").And("bt.delete_at").EqRaw("0").BuildAsSelect().
		Select("jsonb_agg(DISTINCT jsonb_build_object('tagId', bt.tag_id, 'tagName', bt.tag_name, 'color', bt.color)) AS tags").
		Where("ba.article_id").Eq(articleId).
		AndByCondition(!isAdmin, "ba.status").EqRaw("0").
		And("ba.delete_at").EqRaw("0").BuildAsSelect().
		GroupBy("ba.article_id")
	rows, err := self.db.Query(context.Background(), builder.Sql(), articleId)
	defer rows.Close()
	if err == nil && rows.Next() {
		return pgx.RowToAddrOfStructByNameLax[usercase.ArticleVo](rows)
	}
	return nil, err
}

func (self *ArticleRepo) CountByTagId(tagId int) (int64, error) {
	builder := sqlbuild.NewSelectBuilder("t_blog_article").
		Select("count(*)").
		Where(strconv.Itoa(tagId)).EqRaw("ANY(tag_ids)").
		And("delete_at").EqRaw("0").BuildAsSelect()
	row := self.db.QueryRow(context.Background(), builder.Sql())
	var total int64
	err := row.Scan(&total)
	return total, err
}

func (self *ArticleRepo) CountByCategoryId(categoryId int) (int64, error) {
	builder := sqlbuild.NewSelectBuilder("t_blog_article").
		Select("count(*)").
		Where(strconv.Itoa(categoryId)).EqRaw("ANY(category_ids)").
		And("delete_at").EqRaw("0").BuildAsSelect()
	row := self.db.QueryRow(context.Background(), builder.Sql())
	var total int64
	err := row.Scan(&total)
	return total, err
}

func (self *ArticleRepo) CountByTitle(title string, articleId uint64) (uint8, error) {
	builder := sqlbuild.NewSelectBuilder("t_blog_article").
		Select("count(*)").
		Where("article_id").Ne(articleId).
		And("title").Eq(title).
		And("delete_at").EqRaw("0").BuildAsSelect()
	row := self.db.QueryRow(context.Background(), builder.Sql(), builder.Args()...)
	var total uint8
	err := row.Scan(&total)
	return total, err
}

func (self *ArticleRepo) DeleteById(articleId uint64) error {
	builder := sqlbuild.NewUpdateBuilder("t_blog_article").
		Set("delete_at", time.Now().UnixMilli()).
		Where("article_id").Eq(articleId).BuildAsUpdate()
	result, err := self.db.Exec(context.Background(), builder.Sql(), builder.Args()...)
	if err == nil {
		slog.Info("删除博客文章完成", "row", result.RowsAffected(), "articleId", articleId)
	}
	return err
}

func (self *ArticleRepo) VoteUp(articleId uint64, num int) error {
	builder := sqlbuild.NewUpdateBuilder("t_blog_article").
		SetRaw("update_time", "now()").
		SetRaw("vote_up", "vote_up + "+strconv.Itoa(num)).
		Where("article_id").Eq(articleId).BuildAsUpdate()
	result, err := self.db.Exec(context.Background(), builder.Sql(), builder.Args()...)
	if err == nil {
		slog.Info("更新文章点赞数完成", "rows", result.RowsAffected(), "articleId", articleId)
	}
	return err
}

func (self *ArticleRepo) Search(keyword string, limit int) ([]usercase.SimpleArticleVo, error) {
	sql := "select article_id, title, summary from t_blog_article where (title like '%" + keyword + "%' or summary like '%" + keyword + "%') and status = 0 and delete_at = 0 order by sort, view_num desc, create_time desc limit " + strconv.Itoa(limit)
	rows, err := self.db.Query(context.Background(), sql)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (usercase.SimpleArticleVo, error) {
		articleVo := usercase.SimpleArticleVo{}
		scanErr := row.Scan(&articleVo.ArticleId, &articleVo.Title, &articleVo.Summary)
		return articleVo, scanErr
	})
}
